package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
)

func (a *App) viewRegistry(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Services.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	valuesJson, err := json.Marshal(values)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode values")
	}

	viewParams := gin.H{
		"timezone":   a.Config.Timezone,
		"values":     values,
		"valuesJson": string(valuesJson),
	}

	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return nil, fmt.Errorf("unable to get merge reqyests by project, %w", err)
	}

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, fmt.Errorf("unable to get registry by name: %s, %w", registryName, err)
	}

	hasUpdate, _, registryVersion, err := HasUpdate(userCtx, a.Services.Gerrit, reg, MRTargetRegistryVersionUpdate)
	if err != nil {
		return nil, fmt.Errorf("unable to check for updates, %w", err)
	}

	if is, redirect := isVersionRedirect(ctx, "/admin/registry/view", reg.Name, registryVersion); is {
		return redirect, nil
	}

	viewParams["hasUpdate"] = hasUpdate

	for _, f := range a.viewRegistryProcessFunctions(mrs) {
		if err := f(userCtx, reg, values, viewParams); err != nil {
			return nil, errors.Wrap(err, "error during view registry function")
		}
	}

	templateArgs, err := json.Marshal(viewParams)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode template arguments")
	}

	return router.MakeHTMLResponse(200, "registry/view.html", gin.H{
		"page":            "registry",
		"registryVersion": MajorVersion(registryVersion.Core().Original()),
		"templateArgs":    string(templateArgs),
	}), nil
}

func (a *App) viewRegistryProcessFunctions(
	mrs []gerrit.GerritMergeRequest,
) []func(
	ctx context.Context,
	reg *codebase.Codebase,
	values *Values,
	viewParams gin.H,
) error {
	return []func(ctx context.Context, reg *codebase.Codebase, values *Values, viewParams gin.H) error{ // TODO: this is confusing and forces the functions to have the same signature, call the functions directly.
		a.viewRegistryAllowedToEdit,
		a.viewRegistryGetRegistryAndBranches,
		a.viewRegistryGetEDPComponents,
		a.makeViewRegistryGetMergeRequests(mrs),
		a.makeViewRegistryExternalRegistration(mrs),
		a.makeViewRegistryPublicAPI(mrs),
		a.viewDNSConfig,
		a.viewCIDRConfig,
		a.viewAdministratorsConfig,
		a.viewUpdateTrembitaRegistries,
		a.viewGetMasterJobStatus,
		a.viewCreateReleaseJobAvailable,
	}
}

func (a *App) viewGetMasterJobStatus(ctx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	status, _, err := a.Jenkins.GetJobStatus(ctx, fmt.Sprintf("%s/view/MASTER/job/MASTER-Build-%s", reg.Name, reg.Name))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			viewParams["mrAvailable"] = true
			return nil
		}

		return fmt.Errorf("unable to get job status, %w", err)
	}

	viewParams["mrAvailable"] = status == jenkins.StatusSuccess || status == jenkins.StatusNotBuild ||
		status == jenkins.StatusAborted || status == jenkins.StatusFailure

	return nil
}

func (a *App) viewCreateReleaseJobAvailable(ctx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	jobName := fmt.Sprintf("%s/view/Releases/job/Create-release-%s", reg.Name, reg.Name)

	status, _, ststErr := a.Jenkins.GetJobStatus(ctx, jobName)
	isRunning, prgErr := a.Jenkins.IsJobRunning(ctx, jobName)

	handleErr := func(err error) error {
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				viewParams["createReleaseAvailable"] = true
				return nil
			}
			return fmt.Errorf("unable to get job status, %w", err)
		}
		return nil
	}

	if err := handleErr(ststErr); err != nil {
		return err
	}

	if err := handleErr(prgErr); err != nil {
		return err
	}
	viewParams["createReleaseAvailable"] = status == jenkins.StatusSuccess && !isRunning

	return nil
}

func (a *App) viewUpdateTrembitaRegistries(_ context.Context, _ *codebase.Codebase, values *Values, viewParams gin.H) error {
	mrs, ok := viewParams["mergeRequests"]
	if !ok {
		return nil
	}

	trembitaMrs := make(map[string]ExtendedMergeRequests)

	extendedMRs := mrs.([]ExtendedMergeRequests)
	for _, mr := range extendedMRs {
		name, ok := mr.Labels[MRLabelTrembitaRegsitryName]
		if ok {
			trembitaMrs[name] = mr
		}
	}

	for i, r := range values.Trembita.Registries {
		_, ok := trembitaMrs[i]
		if ok {
			r.UserID = "fake"
			values.Trembita.Registries[i] = r
		}
	}

	return nil
}

func (a *App) makeViewRegistryExternalRegistration(mrs []gerrit.GerritMergeRequest) func(userCtx context.Context,
	reg *codebase.Codebase, values *Values, viewParams gin.H) error {
	return func(userCtx context.Context, reg *codebase.Codebase, values *Values, viewParams gin.H) error {
		eRegs, mergeRequestsForER := make([]ExternalRegistration, 0), make(map[string]struct{})

		for _, mr := range mrs {
			if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value == gerrit.StatusNew {
				eRegs = append(eRegs, ExternalRegistration{
					Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
					External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, StatusRegistration: erStatusInactive,
				})
				mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
			} else if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value != gerrit.StatusMerged && mr.Status.Value != gerrit.StatusAbandoned {
				eRegs = append(eRegs, ExternalRegistration{
					Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
					External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, StatusRegistration: erStatusFailed,
				})
				mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
			}
		}

		// TODO: refactor to values struct
		_eRegs, err := decodeExternalRegsFromValues(values.OriginalYaml)
		if err != nil {
			return errors.Wrap(err, "unable to decode external regs")
		}

		for _, _er := range _eRegs {
			if _, ok := mergeRequestsForER[_er.Name]; !ok {
				eRegs = append(eRegs, _er)
			}
		}

		if err := a.loadKeysForExternalRegs(userCtx, reg.Name, eRegs); err != nil {
			return errors.Wrap(err, "unable load keys for ext regs")
		}

		viewParams["externalRegs"] = eRegs
		viewParams["values"] = values

		if err := a.loadCodebasesForExternalRegistrations(reg.Name, eRegs, viewParams); err != nil {
			return errors.Wrap(err, "unable to load codebases for external reg")
		}

		return nil
	}
}

func (a *App) makeViewRegistryPublicAPI(mrs []gerrit.GerritMergeRequest) func(userCtx context.Context,
	reg *codebase.Codebase, values *Values, viewParams gin.H) error {
	return func(userCtx context.Context, reg *codebase.Codebase, values *Values, viewParams gin.H) error {
		publicAPI, mergeRequestsForER := make([]PublicAPI, 0), make(map[string]struct{})
		for _, mr := range mrs {
			if mr.Labels[MRLabelTarget] == mrTargetPublicAPIReg {
				_publicAPI, err := a.makeViewPublicAPIMR(mr, reg.Name)
				if err != nil {
					return fmt.Errorf("unable to get public api from MR, %w", err)
				}
				if _publicAPI.Name != "" {
					publicAPI = append(publicAPI, _publicAPI)
					mergeRequestsForER[_publicAPI.Name] = struct{}{}
				}
			}
		}

		for _, _publicAPI := range values.PublicApi {
			if _, ok := mergeRequestsForER[_publicAPI.Name]; !ok {
				publicAPI = append(publicAPI, _publicAPI)
			}
		}

		viewParams["publicApi"] = publicAPI
		viewParams["values"] = values

		return nil
	}
}

func (a *App) loadKeysForExternalRegs(ctx context.Context, registryName string, eRegs []ExternalRegistration) error {
	for i, er := range eRegs {
		if er.External && er.Enabled {
			s, err := a.Services.K8S.GetSecretFromNamespace(ctx, fmt.Sprintf("keycloak-client-%s-secret", er.Name),
				registryName)
			if k8sErrors.IsNotFound(err) {
				eRegs[i].StatusRegistration = erStatusInactive
				continue
			} else if err != nil {
				return errors.Wrap(err, "unable to get er system key")
			}

			eRegs[i].KeyValue = string(s.Data["clientSecret"])
		}
	}

	return nil
}

func (a *App) loadCodebasesForExternalRegistrations(registryName string, eRegs []ExternalRegistration, viewParams gin.H) error {
	cbs, err := a.Services.Codebase.GetAllByType("registry")
	if err != nil {
		return errors.Wrap(err, "unable to get all registries")
	}

	var availableRegs []codebase.Codebase
	for _, cb := range cbs {
		skip := false
		for _, er := range eRegs {
			if er.Name == cb.Name && !er.External {
				skip = true
				break
			}
		}

		if !skip && cb.Name != registryName && cb.Status.Available && cb.DeletionTimestamp.IsZero() && cb.StrStatus() != "failed" {
			availableRegs = append(availableRegs, cb)
		}
	}

	availableRegsJson, err := json.Marshal(availableRegs)
	if err != nil {
		return errors.Wrap(err, "unable to encode values")
	}

	viewParams["externalRegAvailableRegistriesJSON"] = string(availableRegsJson)

	return nil
}

func convertExternalRegFromInterface(in interface{}) ([]ExternalRegistration, error) {
	js, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode interface to json")
	}

	var res []ExternalRegistration
	if err := json.Unmarshal(js, &res); err != nil {
		return nil, errors.Wrap(err, "unable to decode json")
	}

	return res, nil
}

func (a *App) viewAdministratorsConfig(_ context.Context, _ *codebase.Codebase, values *Values, viewParams gin.H) error {
	viewParams["admins"] = values.Administrators // TODO: remove this
	return nil
}

func (a *App) viewCIDRConfig(userCtx context.Context, reg *codebase.Codebase, values *Values, viewParams gin.H) error {
	if values.Global.WhiteListIP.AdminRoutes != "" {
		viewParams["adminCIDR"] = strings.Split(values.Global.WhiteListIP.AdminRoutes, " ")
	}

	if values.Global.WhiteListIP.CitizenPortal != "" {
		viewParams["citizenCIDR"] = strings.Split(values.Global.WhiteListIP.CitizenPortal, " ")
	}

	if values.Global.WhiteListIP.OfficerPortal != "" {
		viewParams["officerCIDR"] = strings.Split(values.Global.WhiteListIP.OfficerPortal, " ")
	}

	return nil
}

// viewDNSConfig edits the viewParams by reference. The func signature cannot be changed (yet) because it is used in the viewRegistryProcessFunctions.
func (a *App) viewDNSConfig(_ context.Context, _ *codebase.Codebase, values *Values, viewParams gin.H) error {
	// TODO: refactor to values struct
	valuesDict := values.OriginalYaml

	portals, ok := valuesDict["portals"]
	if !ok {
		return nil
	}

	portalsDict := portals.(map[string]interface{})

	if _, ok := portalsDict["citizen"]; ok {
		citizenDict := portalsDict["citizen"].(map[string]interface{})
		if citizenCustomDNS, ok := citizenDict["customDns"]; ok {
			viewParams["citizenPortalHost"] = citizenCustomDNS.(map[string]interface{})["host"].(string)
		}
	}

	if _, ok := portalsDict["officer"]; ok {
		officerDict := portalsDict["officer"].(map[string]interface{})
		if officerCustomDNS, ok := officerDict["customDns"]; ok {
			viewParams["officerPortalHost"] = officerCustomDNS.(map[string]interface{})["host"].(string)
		}
	}

	return nil
}

func (a *App) makeViewRegistryGetMergeRequests(mrs []gerrit.GerritMergeRequest) func(userCtx context.Context,
	reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	return func(userCtx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
		sort.Sort(gerrit.SortByCreationDesc(mrs))

		emrs := make([]ExtendedMergeRequests, 0, len(mrs))
		for _, mr := range mrs {
			if mr.Status.Value == gerrit.StatusNew {
				viewParams["openMergeRequests"] = true
			}
			emrs = append(emrs, ExtendedMergeRequests{GerritMergeRequest: mr})
		}

		viewParams["mergeRequests"] = emrs
		return nil
	}
}

func (a *App) viewRegistryAllowedToEdit(userCtx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := a.Services.Codebase.CheckIsAllowedToUpdate(reg.Name, k8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check codebase creation access")
	}

	viewParams["allowedToEdit"] = allowed
	return nil
}

func (a *App) viewRegistryGetRegistryAndBranches(userCtx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	branches, err := cbService.GetBranchesByCodebase(userCtx, reg.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get registry branches")
	}

	if err := a.loadBranchesStatuses(userCtx, branches); err != nil {
		return errors.Wrap(err, "unable to load branch statuses")
	}

	reg.Branches = branches

	viewParams["registry"] = reg
	viewParams["branches"] = branches
	viewParams["created"] = reg.FormattedCreatedAtTimezone(a.Config.Timezone)

	return nil
}

func (a *App) loadBranchesStatuses(ctx context.Context, branches []codebase.CodebaseBranch) error {
	for i, b := range branches {
		branchName := strings.ToUpper(b.Spec.BranchName)
		status, build, err := a.Jenkins.GetJobStatus(ctx, fmt.Sprintf("%s/view/%s/job/%s-Build-%s", b.Spec.CodebaseName,
			branchName, branchName, b.Spec.CodebaseName))
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				continue
			}

			return errors.Wrap(err, "unable to get branch build status")
		}
		buildString := strconv.FormatInt(build, 10)

		branches[i].Status.Value = status
		branches[i].Status.Build = &buildString
	}

	return nil
}

func (a *App) viewRegistryGetEDPComponents(userCtx context.Context, reg *codebase.Codebase, _ *Values, viewParams gin.H) error {
	jenkinsComponent, err := a.Services.EDPComponent.Get(userCtx, "jenkins")
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.Services.EDPComponent.Get(userCtx, "gerrit")
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit edp component")
	}

	categories, err := a.Services.EDPComponent.GetAllCategory(userCtx, reg.Name)
	if err != nil {
		return errors.Wrap(err, "unable to list namespaced edp components")
	}

	viewParams["jenkinsURL"] = jenkinsComponent.Spec.Url
	viewParams["gerritURL"] = gerritComponent.Spec.Url

	viewParams["platformOperationalComponents"] = categories[edpcomponent.PlatformOperationalZone]
	viewParams["platformAdministrationComponents"] = categories[edpcomponent.PlatformAdministrationZone]

	_, ok := categories[edpcomponent.RegistryOperationalZone]
	if ok {
		viewParams["registryOperationalComponents"] = categories[edpcomponent.RegistryOperationalZone]
		viewParams["registryAdministrationComponents"] = categories[edpcomponent.RegistryAdministrationZone]
	} else {
		// TODO: remove this hotfix
		if err := a.loadRegistryEDPCats(userCtx, reg.Name, viewParams); err != nil {
			return fmt.Errorf("unable to load registry edp component")
		}
	}

	return nil
}

func (a *App) loadRegistryEDPCats(ctx context.Context, registryName string, viewParams gin.H) error {
	components, err := a.Services.EDPComponent.GetAllNamespace(ctx, registryName, true)
	if err != nil {
		return fmt.Errorf("unable to get edp components")
	}

	var (
		registryOperationalComponents    []edpcomponent.EDPComponentItem
		registryAdministrationComponents []edpcomponent.EDPComponentItem
	)

	comMap := map[string]*[]edpcomponent.EDPComponentItem{
		"gerrit":                                 &registryAdministrationComponents,
		"jenkins":                                &registryAdministrationComponents,
		"business-process-administration-portal": &registryAdministrationComponents,
		"citizen-portal":                         &registryOperationalComponents,
		"officer-portal":                         &registryOperationalComponents,
	}

	for _, cmp := range components {
		if list, ok := comMap[cmp.Name]; ok {
			item := edpcomponent.PrepareComponentItem(cmp)
			item.Title = cmp.Name
			*list = append(*list, item)
		}
	}

	viewParams["registryOperationalComponents"] = registryOperationalComponents
	viewParams["registryAdministrationComponents"] = registryAdministrationComponents

	return nil
}
