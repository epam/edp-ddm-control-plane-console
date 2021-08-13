package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type registryWithPermissions struct {
	Registry  *codebase.Codebase
	CanUpdate bool
	CanDelete bool
}

func (r registryWithPermissions) FormattedCreatedAtTimezone(timezone string) string {
	loc, _ := time.LoadLocation(timezone)
	return r.Registry.CreationTimestamp.In(loc).Format(codebase.ViewTimeFormat)
}

func (a *App) listRegistry(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "create", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := a.codebaseService.GetAllByType("registry")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	registries := make([]registryWithPermissions, 0, len(cbs))
	for i := range cbs {
		canGet, err := k8sService.CanI("v2.edp.epam.com", "codebases", "get", cbs[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", cbs[i].Name)
		}
		if !canGet {
			continue
		}

		canUpdate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", cbs[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", cbs[i].Name)
		}

		canDelete, err := k8sService.CanI("v2.edp.epam.com", "codebases", "delete", cbs[i].Name)
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
		"timezone":        a.timezone,
	}), nil
}
