package registry

import (
	"context"
	"encoding/json"
	"sort"

	"ddm-admin-console/service/gerrit"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
)

func (a *App) viewRegistry(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")

	viewParams := gin.H{
		"page":     "registry",
		"timezone": a.timezone,
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
	}
}

func (a *App) viewRegistryExternalRegistration(userCtx context.Context, registryName string, viewParams gin.H) error {
	eRegs, mergeRequestsForER := make([]ExternalRegistration, 0), make(map[string]struct{})
	mrs, err := a.gerritService.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit merge requests")
	}
	for _, mr := range mrs {
		if mr.Labels[mrLabelTarget] == "external-reg" && mr.Status.Value == "NEW" {
			eRegs = append(eRegs, ExternalRegistration{Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
				External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, status: "inactive"})
			mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
		}
	}

	_, _eRegs, err := a.getValuesFromGit(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}
	for _, _er := range _eRegs {
		if _, ok := mergeRequestsForER[_er.Name]; !ok {
			eRegs = append(eRegs, _er)
		}
	}
	viewParams["externalRegs"] = eRegs

	cbs, err := a.codebaseService.GetAllByType("registry")
	if err != nil {
		return errors.Wrap(err, "unable to get all registries")
	}

	var availableRegs []codebase.Codebase
	for _, cb := range cbs {
		skip := false
		for _, er := range eRegs {
			if er.Name == cb.Name {
				skip = true
				break
			}
		}

		if !skip && cb.Name != registryName {
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

func (a *App) viewRegistryGetMergeRequests(userCtx context.Context, registryName string, viewParams gin.H) error {
	mrs, err := a.gerritService.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to list gerrit merge requests")
	}

	sort.Sort(gerrit.SortByCreationDesc(mrs))

	viewParams["mergeRequests"] = mrs
	return nil
}

func (a *App) viewRegistryAllowedToEdit(userCtx context.Context, registryName string, viewParams gin.H) error {
	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := a.codebaseService.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check codebase creation access")
	}

	viewParams["allowedToEdit"] = allowed
	return nil
}

func (a *App) viewRegistryGetRegistryAndBranches(userCtx context.Context, registryName string, viewParams gin.H) error {
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
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
	jenkinsComponent, err := a.edpComponentService.Get(userCtx, "jenkins")
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.edpComponentService.Get(userCtx, "gerrit")
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.edpComponentService.GetAllNamespace(userCtx, registryName, true)
	if err != nil {
		return errors.Wrap(err, "unable to list namespaced edp components")
	}

	viewParams["jenkinsURL"] = jenkinsComponent.Spec.Url
	viewParams["gerritURL"] = gerritComponent.Spec.Url
	viewParams["edpComponents"] = namespacedEDPComponents

	return nil
}