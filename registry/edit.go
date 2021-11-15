package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) editRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")
	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	return router.MakeResponse(200, "registry/edit.html", gin.H{
		"registry":             reg,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"page":                 "registry",
		"hwINITemplateContent": hwINITemplateContent,
	}), nil
}

func (a *App) editRegistryPost(ctx *gin.Context) (response *router.Response, retErr error) {
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
	cb, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	r := registry{
		Name:                registryName,
		RegistryGitBranch:   cb.Spec.DefaultBranch,
		RegistryGitTemplate: cb.Spec.Repository.Url,
		Scenario:            ScenarioKeyNotRequired,
	}

	if err := ctx.ShouldBind(&r); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
	}

	if err := a.editRegistry(&r, ctx.Request, cb, cbService, k8sService); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}

func (a *App) editRegistry(r *registry, rq *http.Request, cb *codebase.Codebase, cbService codebase.ServiceInterface,
	k8sService k8s.ServiceInterface) error {
	if err := a.createRegistryKeys(r, rq, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	cb.Spec.Description = &r.Description
	if cb.Annotations == nil {
		cb.Annotations = make(map[string]string)
	}

	if err := validateAdmins(r.Admins); err != nil {
		return err
	}

	cb.Annotations[AdminsAnnotation] = base64.StdEncoding.EncodeToString([]byte(r.Admins))

	if err := cbService.Update(cb); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	if err := a.jenkinsService.CreateJobBuildRun(fmt.Sprintf("registry-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", r.Name, r.Name), nil); err != nil {
		return errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return nil
}

func validateAdmins(adminsLine string) validator.ValidationErrors {
	validate := validator.New()
	admins := strings.Split(adminsLine, ",")
	for _, admin := range admins {
		errs := validate.Var(admin, "required,email")
		if errs != nil {
			return []validator.FieldError{router.MakeFieldError("Admins", "required")}
		}
	}

	return nil
}
