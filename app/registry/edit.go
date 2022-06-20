package registry

import (
	"context"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
)

func (a *App) editRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")

	if err := a.checkUpdateAccess(registryName, k8sService); err != nil {
		return nil, errors.New("access denied")
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	admins, err := a.admins.GetAdmins(userCtx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry admins")
	}
	model := registry{KeyDeviceType: KeyDeviceTypeFile}
	adminsJs, _ := json.Marshal(admins)
	model.Admins = string(adminsJs)

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	hasUpdate, branches, err := HasUpdate(userCtx, a.Services.Gerrit, reg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	smtpConfig, err := a.getSMTPConfig(userCtx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry SMTP config")
	}

	return router.MakeResponse(200, "registry/edit.html", gin.H{
		"dnsManual":            false,
		"registry":             reg,
		"model":                model,
		"page":                 "registry",
		"hwINITemplateContent": hwINITemplateContent,
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
		"smtpConfig":           smtpConfig,
	}), nil
}

func (a *App) getSMTPConfig(ctx context.Context, registryName string) (map[string]interface{}, error) {
	values, err := a.getValuesFromGit(ctx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	smtpConfig := make(map[string]interface{})
	global, ok := values["global"]
	if !ok {
		return smtpConfig, nil
	}

	globalDict := global.(map[string]interface{})
	emailDict := globalDict["notifications"].(map[string]interface{})["email"].(map[string]interface{})

	return emailDict, nil
}

func (a *App) checkUpdateAccess(codebaseName string, userK8sService k8s.ServiceInterface) error {
	allowedToUpdate, err := a.Services.Codebase.CheckIsAllowedToUpdate(codebaseName, userK8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check create access")
	}
	if !allowedToUpdate {
		return errors.New("access denied")
	}

	return nil
}

func (a *App) editRegistryPost(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")

	allowed, err := cbService.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access")
	}
	if !allowed {
		return nil, errors.New("access denied")
	}

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
		validationErrors, ok := errors.Cause(err).(validator.ValidationErrors)
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

	values, vaultSecretData := make(map[string]interface{}), make(map[string]map[string]interface{})
	if err := a.prepareDNSConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare dns config")
	}

	//if err := a.prepareMailServerConfig(r, vaultSecretData, values); err != nil {
	//	return errors.Wrap(err, "unable to prepare mail server config")
	//}

	if len(vaultSecretData) > 0 {
		if err := a.createVaultSecrets(vaultSecretData); err != nil {
			return errors.Wrap(err, "unable to create vault secrets")
		}
	}

	if len(values) > 0 {
		if err := a.createEditMergeRequest(ginContext, r, values); err != nil {
			return errors.Wrap(err, "unable to create edit merge request")
		}
	}

	admins, err := validateAdmins(r.Admins)
	if err != nil {
		return errors.Wrap(err, "unable to validate admins")
	}
	if err := a.admins.SyncAdmins(ctx, r.Name, admins); err != nil {
		return errors.Wrap(err, "unable to sync admins")
	}

	if err := cbService.Update(cb); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	if err := a.Services.Jenkins.CreateJobBuildRun(fmt.Sprintf("registry-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", r.Name, r.Name), nil); err != nil {
		return errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return nil
}

func (a *App) createEditMergeRequest(ctx *gin.Context, r *registry, globalValues map[string]interface{}) error {
	values, err := a.getValuesFromGit(ctx, r.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}

	globalInterface, ok := values["global"]
	if !ok {
		globalInterface = make(map[string]interface{})
	}

	globalDict, ok := globalInterface.(map[string]interface{})
	if !ok {
		return errors.New("wrong global dict type")
	}

	for k, v := range globalValues {
		globalDict[k] = v
	}

	values["global"] = globalDict

	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return errors.Wrap(err, "unable to encode values yaml")
	}

	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(ctx, r.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get MRs")
	}

	for _, mr := range mrs {
		if mr.Status.Value == "NEW" {
			return MRExists("there is already open merge request(s) for this registry")
		}
	}

	if err := a.Services.Gerrit.CreateMergeRequestWithContents(ctx, &gerrit.MergeRequest{
		ProjectName:   r.Name,
		Name:          fmt.Sprintf("reg-edit-mr-%s-%d", r.Name, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("edit registry"),
		TargetBranch:  "master",
		Labels: map[string]string{
			MRLabelTarget: mrTargetEditRegistry,
		},
		Annotations: map[string]string{},
	}, map[string]string{
		ValuesLocation: string(valuesYaml),
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}

func validateAdmins(adminsLine string) ([]Admin, error) {
	var admins []Admin
	if err := json.Unmarshal([]byte(adminsLine), &admins); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal admins")
	}

	validate := validator.New()
	for i, admin := range admins {
		errs := validate.Var(admin.Email, "required,email")
		if errs != nil {
			return nil,
				validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Admins", "required")})
		}
		admins[i].Username = admin.Email
	}

	return admins, nil
}
