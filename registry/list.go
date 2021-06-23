package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type registryWithPermissions struct {
	Registry  *codebase.Codebase
	CanUpdate bool
	CanDelete bool
}

func (a *App) listRegistry(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := k8sService.CanI("codebase", "create", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := cbService.GetAllByType("registry")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	registries := make([]registryWithPermissions, 0, len(cbs))
	for i := range cbs {
		canUpdate, err := k8sService.CanI("codebase", "update", cbs[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", cbs[i].Name)
		}

		canDelete, err := k8sService.CanI("codebase", "delete", cbs[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", cbs[i].Name)
		}

		registries = append(registries, registryWithPermissions{Registry: &cbs[i], CanDelete: canDelete,
			CanUpdate: canUpdate})
	}

	return router.MakeResponse(200, "registry/list.html", gin.H{
		"registries":      registries,
		"page":            "registry",
		"allowedToCreate": allowedToCreate,
	}), nil
}
