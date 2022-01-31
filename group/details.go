package group

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
)

func (a *App) detailsView(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	groupName := ctx.Param("name")
	group, err := cbService.Get(groupName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get registry by name: %s", groupName)
	}

	branches, err := cbService.GetBranchesByCodebase(groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry branches")
	}

	jenkinsComponent, err := a.edpComponentService.Get("jenkins")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.edpComponentService.Get("gerrit")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.edpComponentService.GetAllNamespace(groupName, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaced edp components")
	}

	allowed, err := a.codebaseService.CheckIsAllowedToUpdate(groupName, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	return router.MakeResponse(200, "group/details.html", gin.H{
		"branches":      branches,
		"group":         group,
		"jenkinsURL":    jenkinsComponent.Spec.Url,
		"gerritURL":     gerritComponent.Spec.Url,
		"page":          "group",
		"edpComponents": namespacedEDPComponents,
		"allowedToEdit": allowed,
		"timezone":      a.timezone,
	}), nil
}
