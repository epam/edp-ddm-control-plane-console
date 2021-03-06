package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (a *App) viewRegistry(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")

	viewParams := gin.H{
		"page":     "registry",
		"timezone": a.Config.Timezone,
	}

	for _, f := range a.viewRegistryProcessFunctions() {
		if err := f(userCtx, registryName, viewParams); err != nil {
			return nil, errors.Wrap(err, "error during view registry function")
		}
	}

	return router.MakeResponse(200, "registry/view.html", viewParams), nil
}

func (a *App) viewRegistryProcessFunctions() []func(ctx context.Context, registryName string, viewParams gin.H) error {
	return []func(ctx context.Context, registryName string, viewParams gin.H) error{
		a.viewRegistryAllowedToEdit,
		a.viewRegistryGetRegistryAndBranches,
		a.viewRegistryGetEDPComponents,
		a.viewRegistryGetMergeRequests,
		a.viewRegistryGetAdmins,
		a.viewRegistryExternalRegistration,
		a.viewDNSConfig,
		a.viewSMTPConfig,
		a.viewCIDRConfig,
	}
}

func (a *App) viewRegistryExternalRegistration(userCtx context.Context, registryName string, viewParams gin.H) error {
	eRegs, mergeRequestsForER := make([]ExternalRegistration, 0), make(map[string]struct{})
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit merge requests")
	}
	for _, mr := range mrs {
		if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value == "NEW" {
			eRegs = append(eRegs, ExternalRegistration{Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
				External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, status: erStatusInactive})
			mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
		} else if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value != "MERGED" && mr.Status.Value != "ABANDONED" {
			eRegs = append(eRegs, ExternalRegistration{Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
				External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, status: erStatusFailed})
			mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
		}
	}

	values, err := a.getValuesFromGit(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}

	_eRegs, err := decodeExternalRegsFromValues(values)
	if err != nil {
		return errors.Wrap(err, "unable to decode external regs")
	}

	for _, _er := range _eRegs {
		if _, ok := mergeRequestsForER[_er.Name]; !ok {
			eRegs = append(eRegs, _er)
		}
	}

	if err := a.loadKeysForExternalRegs(userCtx, registryName, eRegs); err != nil {
		return errors.Wrap(err, "unable load keys for ext regs")
	}

	viewParams["externalRegs"] = eRegs
	viewParams["values"] = values

	if err := a.loadCodebasesForExternalRegistrations(registryName, eRegs, viewParams); err != nil {
		return errors.Wrap(err, "unable to load codebases for external reg")
	}

	return nil
}

func (a *App) loadKeysForExternalRegs(ctx context.Context, registryName string, eRegs []ExternalRegistration) error {
	for i, er := range eRegs {
		if er.External && er.Enabled {
			s, err := a.Services.K8S.GetSecretFromNamespace(ctx, fmt.Sprintf("keycloak-client-%s-secret", er.Name),
				registryName)
			if k8sErrors.IsNotFound(err) {
				eRegs[i].status = erStatusInactive
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

		if !skip && cb.Name != registryName && cb.Status.Available {
			availableRegs = append(availableRegs, cb)
		}
	}
	viewParams["externalRegAvailableRegistries"] = availableRegs

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

func (a *App) viewRegistryGetAdmins(userCtx context.Context, registryName string, viewParams gin.H) error {
	admins, err := a.admins.formatViewAdmins(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to load admins for codebase")
	}

	viewParams["admins"] = admins
	return nil
}

func (a *App) viewGetValues(userCtx context.Context, registryName string, viewParams gin.H) (map[string]interface{}, error) {
	values, ok := viewParams["values"]
	if !ok {
		_values, err := a.getValuesFromGit(userCtx, registryName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get values from git")
		}
		values = _values
	}

	valuesDict, ok := values.(map[string]interface{})
	if !ok {
		return nil, errors.New("wrong values format")
	}

	return valuesDict, nil
}

func (a *App) viewCIDRConfig(userCtx context.Context, registryName string, viewParams gin.H) error {
	valuesDict, err := a.viewGetValues(userCtx, registryName, viewParams)
	if err != nil {
		return errors.Wrap(err, "unable to get values")
	}

	cidr, ok := valuesDict["cidr"]
	if !ok {
		return nil
	}

	cidrDict := cidr.(map[string]interface{})
	if _, ok := cidrDict["admin"]; ok {
		viewParams["adminCIDR"] = cidrDict["admin"].([]interface{})
	}

	if _, ok := cidrDict["citizen"]; ok {
		viewParams["citizenCIDR"] = cidrDict["citizen"].([]interface{})
	}

	if _, ok := cidrDict["officer"]; ok {
		viewParams["officerCIDR"] = cidrDict["officer"].([]interface{})
	}

	return nil
}

func (a *App) viewSMTPConfig(userCtx context.Context, registryName string, viewParams gin.H) error {
	valuesDict, err := a.viewGetValues(userCtx, registryName, viewParams)
	if err != nil {
		return errors.Wrap(err, "unable to get values")
	}

	global := valuesDict["global"].(map[string]interface{})
	notifications, ok := global["notifications"]
	if !ok {
		return nil
	}

	mailType := notifications.(map[string]interface{})["email"].(map[string]interface{})["type"].(string)
	viewParams["smtpType"] = mailType
	return nil
}

func (a *App) viewDNSConfig(userCtx context.Context, registryName string, viewParams gin.H) error {
	valuesDict, err := a.viewGetValues(userCtx, registryName, viewParams)
	if err != nil {
		return errors.Wrap(err, "unable to get values")
	}

	portals, ok := valuesDict["portals"]
	if !ok {
		return nil
	}

	portalsDict := portals.(map[string]interface{})
	citizenDict := portalsDict["citizen"].(map[string]interface{})
	officerDict := portalsDict["citizen"].(map[string]interface{})

	if citizenCustomDNS, ok := citizenDict["customDns"]; ok {
		viewParams["citizenPortalHost"] = citizenCustomDNS.(map[string]interface{})["host"].(string)
	}

	if officerCustomDNS, ok := officerDict["customDns"]; ok {
		viewParams["officerPortalHost"] = officerCustomDNS.(map[string]interface{})["host"].(string)
	}

	return nil
}

func (a *App) viewRegistryGetMergeRequests(userCtx context.Context, registryName string, viewParams gin.H) error {
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to list gerrit merge requests")
	}

	sort.Sort(gerrit.SortByCreationDesc(mrs))

	emrs := make([]ExtendedMergeRequests, 0, len(mrs))
	for _, mr := range mrs {
		emrs = append(emrs, ExtendedMergeRequests{GerritMergeRequest: mr})
	}

	viewParams["mergeRequests"] = emrs
	return nil
}

func (a *App) viewRegistryAllowedToEdit(userCtx context.Context, registryName string, viewParams gin.H) error {
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := a.Services.Codebase.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check codebase creation access")
	}

	viewParams["allowedToEdit"] = allowed
	return nil
}

func (a *App) viewRegistryGetRegistryAndBranches(userCtx context.Context, registryName string, viewParams gin.H) error {
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	registry, err := cbService.Get(registryName)
	if err != nil {
		return errors.Wrapf(err, "unable to get registry by name: %s", registryName)
	}

	branches, err := cbService.GetBranchesByCodebase(registry.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get registry branches")
	}
	registry.Branches = branches

	viewParams["registry"] = registry
	viewParams["branches"] = branches

	return nil
}

func (a *App) viewRegistryGetEDPComponents(userCtx context.Context, registryName string, viewParams gin.H) error {
	jenkinsComponent, err := a.Services.EDPComponent.Get(userCtx, "jenkins")
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.Services.EDPComponent.Get(userCtx, "gerrit")
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.Services.EDPComponent.GetAllNamespace(userCtx, registryName, true)
	if err != nil {
		return errors.Wrap(err, "unable to list namespaced edp components")
	}

	viewParams["jenkinsURL"] = jenkinsComponent.Spec.Url
	viewParams["gerritURL"] = gerritComponent.Spec.Url
	viewParams["edpComponents"] = namespacedEDPComponents

	return nil
}
