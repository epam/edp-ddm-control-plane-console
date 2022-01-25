package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ddm-admin-console/service/gerrit"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"

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

	gerritProject, err := a.gerritService.GetProject(userCtx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit project")
	}

	hasUpdate, branches, err := a.hasUpdate(userCtx, gerritProject)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	return router.MakeResponse(200, "registry/edit.html", gin.H{
		"registry":             reg,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"page":                 "registry",
		"hwINITemplateContent": hwINITemplateContent,
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
	}), nil
}

func (a *App) hasUpdate(ctx context.Context, gerritProject *gerrit.GerritProject) (bool, []string, error) {
	branches := updateBranches(gerritProject.Status.Branches)

	if len(branches) == 0 {
		return false, branches, nil
	}

	mrs, err := a.gerritService.GetMergeRequestByProject(ctx, gerritProject.Spec.Name)
	if err != nil {
		return false, branches, errors.Wrap(err, "unable to get merge requests")
	}

	branchesDict := make(map[string]string)
	for _, br := range branches {
		branchesDict[br] = br
	}

	for _, mr := range mrs {
		if mr.Status.Value == "NEW" {
			return false, branches, nil
		}

		if mr.Status.Value == "MERGED" {
			delete(branchesDict, mr.Spec.SourceBranch)
		}
	}

	branches = []string{}
	for _, br := range branchesDict {
		branches = append(branches, br)
	}

	return true, branches, nil
}

func updateBranches(projectBranches []string) []string {
	var updateBranches []string
	for _, br := range projectBranches {
		if strings.Contains(br, "refs/heads") && !strings.Contains(br, "master") {
			updateBranches = append(updateBranches, strings.Replace(br, "refs/heads/", "", 1))
		}
	}

	return updateBranches
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

	if err := a.editRegistry(userCtx, ctx, &r, cb, cbService, k8sService); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", r.Name)), nil
}

func (a *App) editRegistry(ctx context.Context, ginContext *gin.Context, r *registry, cb *codebase.Codebase,
	cbService codebase.ServiceInterface, k8sService k8s.ServiceInterface) error {
	if err := a.createRegistryKeys(r, ginContext.Request, k8sService); err != nil {
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

	if r.UpdateBranch != "" {
		prj, err := a.gerritService.GetProject(ctx, r.Name)
		if err != nil {
			return errors.Wrap(err, "unable to get registry gerrit project")
		}

		if err := a.gerritService.CreateMergeRequest(ctx, &gerrit.MergeRequest{
			CommitMessage: fmt.Sprintf("Update registry to %s", r.UpdateBranch),
			SourceBranch:  r.UpdateBranch,
			ProjectName:   prj.Spec.Name,
			Name:          fmt.Sprintf("%s-update-%d", r.Name, time.Now().Unix()),
			AuthorName:    ginContext.GetString(router.UserNameSessionKey),
			AuthorEmail:   ginContext.GetString(router.UserEmailSessionKey),
		}); err != nil {
			return errors.Wrap(err, "unable to create update merge request")
		}
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
