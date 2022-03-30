package registry

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
)

func (a *App) viewRegistry(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")
	registry, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get registry by name: %s", registryName)
	}

	branches, err := cbService.GetBranchesByCodebase(registry.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry branches")
	}

	registry.Branches = branches
	jenkinsComponent, err := a.edpComponentService.Get("jenkins")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.edpComponentService.Get("gerrit")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.edpComponentService.GetAllNamespace(registry.Name, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaced edp components")
	}

	mrs, err := a.gerritService.GetMergeRequestByProject(ctx, registry.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit merge requests")
	}

	allowed, err := a.codebaseService.CheckIsAllowedToUpdate(registry.Name, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	admins, err := a.formatViewAdmins(userCtx, registry.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load admins for codebase")
	}

	return router.MakeResponse(200, "registry/view.html", gin.H{
		"branches":      branches,
		"registry":      registry,
		"jenkinsURL":    jenkinsComponent.Spec.Url,
		"gerritURL":     gerritComponent.Spec.Url,
		"page":          "registry",
		"edpComponents": namespacedEDPComponents,
		"allowedToEdit": allowed,
		"mergeRequests": mrs,
		"timezone":      a.timezone,
		"admins":        admins,
	}), nil
}

func (a *App) formatViewAdmins(ctx context.Context, registryName string) (string, error) {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return "", errors.Wrap(err, "unable to load admins")
	}

	var registryAdmins []string

	for _, u := range usrs {
		for _, role := range u.Spec.Roles {
			if role == groupRoleName(registryName) {
				registryAdmins = append(registryAdmins, u.Spec.Email)
				break
			}
		}
	}

	return strings.Join(registryAdmins, ", "), nil
}
