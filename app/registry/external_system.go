package registry

import (
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	authTypeNoAuth          = "NO_AUTH"
	authTypeAuthToken       = "AUTH_TOKEN"
	authTypeBearer          = "BEARER"
	authTypeBasic           = "BASIC"
	authTypeAuthTokenBearer = "AUTH_TOKEN+BEARER"
	registryNamePlaceholder = "{NAME_REGISTRY}"
)

type RegistryExternalSystemForm struct {
	RegistryName        string `form:"external-system-registry-name" binding:"required"`
	URL                 string `form:"external-system-url"`
	Protocol            string `form:"external-system-protocol" binding:"required"`
	AuthType            string `form:"external-system-auth-type" binding:"required"`
	AuthURL             string `form:"external-system-auth-url"`
	AccessTokenJSONPath string `form:"external-system-auth-access-token-json-path"`
	AuthSecret          string `form:"external-system-auth-secret"`
	AuthUsername        string `form:"external-system-auth-username"`
}

func externalSystemsSecretPath(vaultRegistryPath string, exRegistryName string) string {
	return fmt.Sprintf("%s/external-systems/%s-%s", vaultRegistryPath, exRegistryName, time.Now().Format("20060201T150405Z"))
}

func externalSystemSecretPrefixedPath(originalPath string) string {
	return fmt.Sprintf("vault:%s", originalPath)
}

func (f RegistryExternalSystemForm) ToNestedForm(vaultRegistryPath, wiremockAddr string) ExternalSystem {
	es := ExternalSystem{
		URL:      f.URL,
		Protocol: f.Protocol,
		Auth: map[string]string{
			"type": f.AuthType,
		},
	}

	if f.AuthType != authTypeNoAuth {
		es.Auth["secret"] = externalSystemSecretPrefixedPath(vaultRegistryPath)
	}

	if f.AuthType == authTypeAuthTokenBearer {
		es.Auth["auth-url"] = f.AuthURL
		es.Auth["access-token-json-path"] = f.AccessTokenJSONPath
	}

	if es.URL == "" {
		es.Mock = true
		es.URL = wiremockAddr
	}

	return es
}

func (a *App) getBasicUsername(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	systemRegsitryName := ctx.Query("registry-name")
	if systemRegsitryName == "" {
		return nil, errors.New("bad request")
	}

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	cbService, err := a.Services.Codebase.ServiceForContext(router.ContextWithUserAccessToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	_, err = cbService.Get(registryName)
	if err != nil {
		return nil, fmt.Errorf("unable to find registry, %w", err)
	}
	secret := values.ExternalSystems[systemRegsitryName].Auth["secret"]

	dataDict, err := a.Vault.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("unable to load id-gov-ua secret, err: %w", err)
	}

	d, ok := dataDict[fmt.Sprintf("external-systems.%s.auth.secret.username", systemRegsitryName)]
	if !ok {
		return nil, errors.New("no basic data")
	}

	str, ok := d.(string)
	if !ok {
		return nil, errors.New("wrong basic data")
	}

	return router.MakeJSONResponse(200, str), nil
}

func (a *App) deleteExternalSystem(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	exSystemName := ctx.Query("external-system")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	eSys, ok := values.ExternalSystems[exSystemName]
	if !ok {
		return nil, errors.New("external system does not exists")
	}

	if eSys.Type == "platform" {
		return nil, errors.New("external system is unavailable to delete")
	}

	delete(values.ExternalSystems, exSystemName)

	values.OriginalYaml[externalSystemsKey] = values.ExternalSystems

	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit, []string{}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) checkExternalSystemExists(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	exSystemName := ctx.Query("external-system")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	_, ok := values.ExternalSystems[exSystemName]
	if ok {
		return router.MakeStatusResponse(http.StatusOK), nil
	}

	return router.MakeStatusResponse(http.StatusNotFound), nil
}

func (a *App) createExternalSystemRegistry(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	var f RegistryExternalSystemForm
	if err := ctx.ShouldBind(&f); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	_, ok := values.ExternalSystems[f.RegistryName]
	if ok {
		return nil, errors.Wrap(err, "external system already exists")
	}

	var secretPath = externalSystemsSecretPath(a.vaultRegistryPath(registryName), f.RegistryName)
	extenalSystem := f.ToNestedForm(secretPath, strings.ReplaceAll(a.WiremockAddr, registryNamePlaceholder, registryName))
	extenalSystem.Protocol = externalSystemDefaultProtocol
	extenalSystem.Type = externalSystemDeletableType

	valuesExternalSystems, ok := values.OriginalYaml[externalSystemsKey]
	if !ok {
		return nil, errors.Wrap(err, "no external systems key in values")
	}
	valuesExternalSystemsDict := valuesExternalSystems.(map[string]interface{})

	valuesExternalSystemsDict[f.RegistryName] = extenalSystem
	values.OriginalYaml[externalSystemsKey] = valuesExternalSystemsDict

	if err := a.setExternalSystemRegistrySecrets(&f, secretPath); err != nil {
		return nil, errors.Wrap(err, "unable to set external system")
	}

	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit, []string{}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

// edit
func (a *App) setExternalSystemRegistryData(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	var f RegistryExternalSystemForm

	if err := ctx.ShouldBind(&f); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}

	var secretPath = externalSystemsSecretPath(a.vaultRegistryPath(registryName), f.RegistryName)

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	valuesExternalSystem, ok := values.ExternalSystems[f.RegistryName]
	if !ok {
		return nil, errors.Wrap(err, "unable to get external system")
	}

	editExtenalSystem := f.ToNestedForm(secretPath,
		strings.ReplaceAll(a.WiremockAddr, registryNamePlaceholder, registryName))
	editExtenalSystem.Type = valuesExternalSystem.Type
	editExtenalSystem.Protocol = valuesExternalSystem.Protocol

	valuesExternalSystems, ok := values.OriginalYaml[externalSystemsKey]
	if !ok {
		return nil, errors.Wrap(err, "no external systems key in values")
	}
	valuesExternalSystemsDict := valuesExternalSystems.(map[string]interface{})

	valuesExternalSystemsDict[f.RegistryName] = editExtenalSystem
	values.OriginalYaml[externalSystemsKey] = valuesExternalSystemsDict

	if err := a.setExternalSystemRegistrySecrets(&f, secretPath); err != nil {
		return nil, errors.Wrap(err, "unable to set external system")
	}

	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit, []string{}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) setExternalSystemRegistrySecrets(f *RegistryExternalSystemForm, secretPath string) error {
	secretData := make(map[string]interface{})
	prefixedPath := externalSystemSecretPrefixedPath(secretPath)

	createSecrets := false

	if f.AuthType == authTypeAuthToken || f.AuthType == authTypeAuthTokenBearer || f.AuthType == authTypeBearer {
		if f.AuthSecret == prefixedPath {
			return nil
		}

		secretData[fmt.Sprintf("external-systems.%s.auth.secret.token", f.RegistryName)] = f.AuthSecret
		createSecrets = true
	} else if f.AuthType == authTypeBasic {

		if f.AuthUsername != prefixedPath {
			secretData[fmt.Sprintf("external-systems.%s.auth.secret.username", f.RegistryName)] = f.AuthUsername
			createSecrets = true
		}

		if f.AuthSecret != prefixedPath {
			secretData[fmt.Sprintf("external-systems.%s.auth.secret.password", f.RegistryName)] = f.AuthSecret
			createSecrets = true
		}
	}

	if !createSecrets {
		return nil
	}

	if err := CreateVaultSecrets(a.Vault, map[string]map[string]interface{}{
		secretPath: secretData,
	}, true); err != nil {
		return errors.Wrap(err, "unable to create auth token secret")
	}

	return nil
}
