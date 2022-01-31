package group

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/registry"

	"ddm-admin-console/service/codebase"

	"ddm-admin-console/router"
)

type listView struct {
	app *App
}

type groupWithRegistries struct {
	Group      *codebase.WithPermissions
	Registries []codebase.Codebase
}

func (l listView) Get(ctx *gin.Context) (*router.Response, error) {
	userCtx := l.app.router.ContextWithUserAccessToken(ctx)
	k8sService, err := l.app.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := l.app.codebaseService.CheckIsAllowedToCreate(k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := l.app.codebaseService.GetAllByType(codebase.GroupCodebaseType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	groups, err := l.app.codebaseService.CheckPermissions(cbs, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check permissions for groups")
	}

	groupsWithRegistries, err := l.prepareGroupRegistries(groups)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare groups with registries")
	}

	return router.MakeResponse(200, "group/list.html", gin.H{
		"groups":          groupsWithRegistries,
		"page":            "group",
		"allowedToCreate": allowedToCreate,
	}), nil
}

func (l listView) prepareGroupRegistries(groups []codebase.WithPermissions) ([]groupWithRegistries, error) {
	groupsDict := make(map[string]groupWithRegistries)
	for i, gr := range groups {
		groupsDict[gr.Codebase.Name] = groupWithRegistries{Group: &groups[i]}
	}

	registries, err := l.app.codebaseService.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list registries")
	}

	for _, reg := range registries {
		regGroup, ok := reg.Annotations[registry.GroupAnnotation]
		if ok {
			regGroupBytes, err := base64.StdEncoding.DecodeString(regGroup)
			if err != nil {
				return nil, errors.Wrap(err, "unable to decode registry group")
			}

			regGroupStr := strings.TrimSpace(string(regGroupBytes))

			if group, ok := groupsDict[regGroupStr]; ok {
				group.Registries = append(group.Registries, reg)
				groupsDict[regGroupStr] = group
			}
		}
	}

	result := make([]groupWithRegistries, 0, len(groupsDict))
	for _, v := range groupsDict {
		result = append(result, v)
	}

	return result, nil
}

func (l listView) Post(ctx *gin.Context) (*router.Response, error) {
	userCtx := l.app.router.ContextWithUserAccessToken(ctx)
	cbService, err := l.app.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	groupName := ctx.PostForm("registry-name")

	hasRegs, err := l.groupHasRegistries(groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check group registries")
	}
	if hasRegs {
		return nil, errors.New("group has registries")
	}

	err = cbService.Delete(groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/group/overview"), nil
}

func (l listView) groupHasRegistries(groupName string) (bool, error) {
	registries, err := l.app.codebaseService.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return false, errors.Wrap(err, "unable to list registries")
	}

	for _, reg := range registries {
		regGroup, ok := reg.Annotations[registry.GroupAnnotation]
		if ok {
			regGroupBytes, err := base64.StdEncoding.DecodeString(regGroup)
			if err != nil {
				return false, errors.Wrap(err, "unable to decode registry group")
			}

			if groupName == strings.TrimSpace(string(regGroupBytes)) {
				return true, nil
			}
		}
	}

	return false, nil
}
