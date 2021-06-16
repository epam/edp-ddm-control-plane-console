package registry

import (
	"ddm-admin-console/router"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

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

	allowedToCreate, err := k8sService.CanCreateCodebase()
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := cbService.GetAllByType("registry")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	return router.MakeResponse(200, "registry/list.html", gin.H{
		"registries":      cbs,
		"page":            "registry",
		"allowedToCreate": allowedToCreate,
	}), nil
}
