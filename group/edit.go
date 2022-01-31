package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
)

type editView struct {
	app *App
}

func (e editView) Get(ctx *gin.Context) (*router.Response, error) {
	userCtx := e.app.router.ContextWithUserAccessToken(ctx)
	groupName := ctx.Param("name")

	if err := e.validateAccess(groupName, userCtx); err != nil {
		return nil, errors.Wrap(err, "unable to validate access")
	}

	cbService, err := e.app.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	gr, err := cbService.Get(groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	return router.MakeResponse(200, "group/edit.html", gin.H{
		"group": gr,
		"page":  "group",
	}), nil
}

func (e editView) Post(ctx *gin.Context) (*router.Response, error) {
	groupName := ctx.Param("name")
	userCtx := e.app.router.ContextWithUserAccessToken(ctx)

	if err := e.validateAccess(groupName, userCtx); err != nil {
		return nil, errors.Wrap(err, "unable to validate access")
	}

	gr := group{
		Name: groupName,
	}

	if err := ctx.ShouldBind(&gr); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "group/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "group": gr}), nil
	}

	cbService, err := e.app.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	groupCodebase, err := cbService.Get(groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}
	groupCodebase.Spec.Description = &gr.Description

	if err := cbService.Update(groupCodebase); err != nil {
		return nil, errors.Wrap(err, "unable to update codebase")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/group/view/%s", groupName)), nil
}

func (e editView) validateAccess(groupName string, userCtx context.Context) error {
	k8sService, err := e.app.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := e.app.codebaseService.CheckIsAllowedToUpdate(groupName, k8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check access")

	}

	if !allowed {
		return errors.New("access denied")
	}

	return nil
}
