package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/vault"
)

const (
	AnnotationSMPTType             = "registry-parameters/smtp-type"
	AnnotationSMPTOpts             = "registry-parameters/smtp-opts"
	AnnotationTemplateName         = "registry-parameters/template-name"
	AnnotationCreatorUsername      = "registry-parameters/creator-username"
	AnnotationCreatorEmail         = "registry-parameters/creator-email"
	AnnotationValues               = "registry-parameters/values"
	AnnotationTemplatePushedStatus = "registry-status/template-pushed"
	AdministratorsValuesKey        = "administrators"
	ResourcesValuesKey             = "registry"
	VaultKeyCACert                 = "caCertificate"
	VaultKeyCert                   = "certificate"
	VaultKeyPK                     = "key"
	externalSystemsKey             = "external-systems"
	externalSystemDefaultProtocol  = "REST"
	externalSystemDeletableType    = "registry"
	trembitaRegistriesKey          = "registries"
	trembitaValuesKey              = "trembita"
	trembitaRegistriesValuesKet    = "registries"
)

type processorFunction func(
	ctx *gin.Context, // TODO: it is generally a bad practise to pass the raw request down the program.
	r *registry,
	values *Values,
	secrets map[string]map[string]any,
	mrActions *[]string,
) (
	bool,
	error,
)

func (a *App) createUpdateRegistryProcessors() []processorFunction {
	return []processorFunction{
		a.prepareDNSConfig,
		a.prepareCIDRConfig,
		a.prepareMailServerConfig,
		a.prepareAdminsConfig,
		a.prepareRegistryResources,
		a.prepareSupplierAuthConfig,
		a.prepareBackupSchedule,
		a.prepareKeycloakCustomHostname,
		a.prepareCitizenAuthSettings,
		a.prepareTrembitaIPList,
		a.prepareDigitalDocuments,
		a.prepareGriada,
		a.prepareGlobalValuesYaml,
	}
}

func (a *App) validatePEMFile(ctx *gin.Context) (rsp router.Response, retErr error) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get form file")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file data")
	}

	if _, err := DecodePEM(data); err != nil {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	return router.MakeStatusResponse(http.StatusOK), nil
}

func (a *App) registryNameAvailable(ctx *gin.Context) (rsp router.Response, retErr error) {
	name := ctx.Param("name")
	_, err := a.Codebase.Get(name)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return router.MakeJSONResponse(http.StatusOK, gin.H{"registryNameAvailable": true}), nil
		}

		return nil, errors.Wrap(err, "unable to check codebase existance")
	}

	return router.MakeJSONResponse(http.StatusOK, gin.H{"registryNameAvailable": false}), nil
}

func (a *App) createRegistryGet(ctx *gin.Context) (response router.Response, retErr error) {
	prjs, err := a.Services.Gerrit.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs, a.Config.RegistryTemplateName)

	userCtx := router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	hwINITemplateContent, err := GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	dnsManual, err := a.getDNSManualURL(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dns manual")
	}

	gerritBranches := formatGerritProjectBranches(prjs)

	keycloakHostname, err := LoadKeycloakDefaultHostname(ctx, a.KeycloakDefaultHostname, a.EDPComponent)
	if err != nil {
		return nil, fmt.Errorf("unable to load keycloak default hostname, %w", err)
	}

	keycloakHostnames, err := a.loadKeycloakHostnames()
	if err != nil {
		return nil, fmt.Errorf("unable to load keycloak hostnames, %w", err)
	}

	clusterValues, err := a.getValuesFromBranch(a.ClusterCodebaseName, MasterBranch)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster values, %w", err)
	}

	responseParams := gin.H{
		"dnsManual":            dnsManual,
		"page":                 "registry",
		"gerritProjects":       prjs,
		"gerritBranches":       gerritBranches,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"smtpConfig":           "{}",
		"action":               "create",
		"registryData":         "{}",
		"keycloakHostname":     keycloakHostname,
		"keycloakHostnames":    keycloakHostnames,
		"registryTemplateName": a.Config.RegistryTemplateName,
		"platformStatusType":   a.Config.CloudProvider,
		"registryVersion":      ctx.Request.URL.Query().Get("version"),
		"clusterValues":        clusterValues,
	}

	templateArgs, err := json.Marshal(responseParams)
	if err != nil {
		return nil, fmt.Errorf("failed to encode template arguments: %w", err)
	}

	responseParams["templateArgs"] = string(templateArgs)
	responseParams["hwINITemplateContent"] = hwINITemplateContent

	return router.MakeHTMLResponse(200, "registry/create.html", responseParams), nil
}

func GetManualURL(ctx context.Context, edpComponentService edpComponent.ServiceInterface, ddmManualComponent, manualPath string) (string, error) {
	com, err := edpComponentService.Get(ctx, ddmManualComponent)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return "", nil
		}

		return "", fmt.Errorf("unable to get edp component, %w", err)
	}

	u, err := url.Parse(com.Spec.Url)
	if err != nil {
		return "", fmt.Errorf("unable to parse url, %w", err)
	}

	if strings.Contains(manualPath, "#") {
		parts := strings.Split(manualPath, "#")
		manualPath = parts[0]
		u.Path = path.Join(u.Path, manualPath)
		return fmt.Sprintf("%s#%s", u.String(), parts[1]), nil
	}

	u.Path = path.Join(u.Path, manualPath)
	return u.String(), nil
}

func (a *App) getDNSManualURL(ctx context.Context) (string, error) {
	return GetManualURL(ctx, a.EDPComponent, a.Config.DDMManualEDPComponent, a.Config.RegistryDNSManualPath)
}

func headsCount(refs []string) int {
	cnt := 0
	for _, ref := range refs {
		if strings.Contains(ref, "refs/heads") {
			cnt += 1
		}
	}

	return cnt
}

func (a *App) filterProjects(projects []gerrit.GerritProject, prjName string) []gerrit.GerritProject {
	filteredProjects := make([]gerrit.GerritProject, 0, 4)
	for _, prj := range projects {
		if prj.Spec.Name == prjName {
			if headsCount(prj.Status.Branches) > 1 {
				var branches []string
				for _, br := range prj.Status.Branches {
					if !strings.Contains(br, "master") {
						branches = append(branches, br)
					}
				}
				prj.Status.Branches = branches
			}

			filteredProjects = append(filteredProjects, prj)
		}
	}

	return filteredProjects
}

func formatGerritProjectBranches(projects []gerrit.GerritProject) []string {
	var branches []string
	for _, p := range projects {
		for _, b := range p.Status.Branches {
			idx := strings.Index(b, "heads/")
			if idx != -1 && !slices.Contains(branches, b[idx+6:]) {
				branches = append(branches, b[idx+6:])
			}
		}
	}
	return branches
}

func GetINITemplateContent(path string) (string, error) {
	iniTemplate, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "unable to open ini template file")
	}

	data, err := ioutil.ReadAll(iniTemplate)
	if err != nil {
		return "", errors.Wrap(err, "unable to read ini template data")
	}

	return string(data), nil
}

func (a *App) createRegistryPost(ctx *gin.Context) (response router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get k8s service for user")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	r := registry{Scenario: ScenarioKeyNotRequired}
	if a.Config.Region == "ua" {
		r = registry{Scenario: ScenarioKeyRequired}
	}

	if err := ctx.ShouldBind(&r); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}

	if err := a.createRegistry(userCtx, ctx, &r, cbService); err != nil {
		return nil, errors.Wrap(err, "unable to create registry")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}

func (a *App) checkCreateAccess(userK8sService k8s.ServiceInterface) error {
	allowedToCreate, err := a.Services.Codebase.CheckIsAllowedToCreate(userK8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check create access")
	}
	if !allowedToCreate {
		return errors.New("access denied")
	}

	return nil
}

func (a *App) createRegistry(
	ctx context.Context,
	ginContext *gin.Context,
	newRegistry *registry,
	cbService codebase.ServiceInterface,
) error {
	_, err := cbService.Get(newRegistry.Name)
	if err == nil {
		return validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Name", "registry-exists")})
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unknown error")
	}

	registryTemplate, err := GetValuesFromGit(a.Config.RegistryTemplateName, newRegistry.RegistryGitBranch, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to load registryTemplate from template")
	}

	vaultSecretData := make(map[string]map[string]interface{})
	mrActions := make([]string, 0)

	for _, proc := range a.createUpdateRegistryProcessors() {
		if _, err := proc(ginContext, newRegistry, registryTemplate, vaultSecretData, &mrActions); err != nil {
			return errors.Wrap(err, "error during registry create")
		}
	}

	repoFiles := make(map[string]string)

	if _, err := PrepareRegistryKeys(
		keyManagement{
			r: newRegistry,
			vaultSecretPath: a.vaultRegistryPathKey(
				newRegistry.Name,
				fmt.Sprintf("%s-%s", KeyManagementVaultPath, time.Now().Format("20060201T150405Z")),
			),
		},
		ginContext.Request,
		vaultSecretData,
		registryTemplate,
		repoFiles,
		a.Services.Vault,
	); err != nil {
		return errors.Wrap(err, "unable to prepare registry keys")
	}

	if err := CacheRepoFiles(a.TempFolder, newRegistry.Name, repoFiles, a.Cache); err != nil {
		return fmt.Errorf("unable to cache repo file, %w", err)
	}

	if err := CreateVaultSecrets(a.Vault, vaultSecretData, false); err != nil {
		return errors.Wrap(err, "unable to create vault secrets")
	}

	if err := a.Services.Gerrit.CreateProject(ctx, newRegistry.Name); err != nil {
		return errors.Wrap(err, "unable to create gerrit project")
	}

	cb := a.prepareRegistryCodebase(newRegistry)
	valuesEncoded, err := a.encodeValues(newRegistry.Name, registryTemplate.OriginalYaml)
	if err != nil {
		return errors.Wrap(err, "unable to encode values")
	}

	cb.Annotations = map[string]string{
		AnnotationTemplateName:         a.Config.RegistryTemplateName,
		AnnotationCreatorUsername:      ginContext.GetString(router.UserNameSessionKey),
		AnnotationCreatorEmail:         ginContext.GetString(router.UserEmailSessionKey),
		AnnotationValues:               valuesEncoded,
		AnnotationTemplatePushedStatus: "false",
	}

	if err := cbService.Create(cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := cbService.CreateDefaultBranch(cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func CreateVaultSecrets(v vault.ServiceInterface, secretData map[string]map[string]interface{}, append bool) error {
	for vPath, pathSecretData := range secretData {
		if append {
			sec, err := v.Read(vPath)
			if err != nil && !errors.Is(err, vault.ErrSecretIsNil) {
				return errors.Wrap(err, "unable to read secret")
			}

			if errors.Is(err, vault.ErrSecretIsNil) || sec == nil {
				sec = make(map[string]interface{})
			}

			for k, v := range pathSecretData {
				sec[k] = v
			}

			pathSecretData = sec
		}

		if _, err := v.Write(vPath, pathSecretData); err != nil {
			return errors.Wrap(err, "unable to write to vault")
		}
	}

	return nil
}

func (a *App) encodeValues(registryName string, values map[string]interface{}) (string, error) {
	values["registryVaultPath"] = a.vaultRegistryPath(registryName)
	bts, err := json.Marshal(values)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode values to JSON")
	}

	return string(bts), nil
}

func (a *App) vaultRegistryPath(registryName string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(a.Config.VaultRegistrySecretPathTemplate, "{registry}", registryName),
		"{engine}", a.Config.VaultKVEngineName)
}

func (a *App) vaultRegistryPathKey(registryName, key string) string {
	return fmt.Sprintf("%s/%s", a.vaultRegistryPath(registryName), key)
}

func (a *App) keyManagementRegistryVaultPath(registryName string) string {
	return a.vaultRegistryPath(registryName) + "/key-management"
}

func (a *App) prepareCIDRConfig(ctx *gin.Context, r *registry, _values *Values,
	_ map[string]map[string]interface{}, _ *[]string,
) (bool, error) {
	if ctx.PostForm("action") == "edit" && ctx.PostForm("cidr-changed") == "" {
		return false, nil
	}

	globalInterface, ok := _values.OriginalYaml[GlobalValuesIndex]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	if err := handleCIDRCategory(r.CIDRCitizen, &_values.Global.WhiteListIP.CitizenPortal); err != nil {
		return false, errors.Wrap(err, "unable to handle cidr category")
	}

	if err := handleCIDRCategory(r.CIDROfficer, &_values.Global.WhiteListIP.OfficerPortal); err != nil {
		return false, errors.Wrap(err, "unable to handle cidr category")
	}

	if err := handleCIDRCategory(r.CIDRAdmin, &_values.Global.WhiteListIP.AdminRoutes); err != nil {
		return false, errors.Wrap(err, "unable to handle cidr category")
	}

	globalDict[WhiteListIPIndex] = _values.Global.WhiteListIP
	_values.OriginalYaml[GlobalValuesIndex] = globalDict

	return true, nil
}

func handleCIDRCategory(cidrCategory string, categoryValue *string) error {
	if cidrCategory == "" {
		return nil
	}

	var cidr []string
	if err := json.Unmarshal([]byte(cidrCategory), &cidr); err != nil {
		return errors.Wrap(err, "unable to decode cidr")
	}

	*categoryValue = strings.Join(cidr, " ")

	return nil
}

func (a *App) prepareAdminsConfig(_ *gin.Context, r *registry, _values *Values,
	secrets map[string]map[string]interface{}, _ *[]string,
) (bool, error) {
	values := _values.OriginalYaml
	// TODO: refactor to new values

	// TODO: don't recreate admin secrets for existing admin
	if r.Admins != "" && r.AdminsChanged == "on" {
		admins, err := validateAdmins(r.Admins)
		if err != nil {
			return false, errors.Wrap(err, "unable to validate admins")
		}

		adminsVaultPath := a.vaultRegistryPathKey(r.Name, "administrators")
		for i, adm := range admins {
			adminVaultPath := fmt.Sprintf("%s/%s", adminsVaultPath, adm.Email)
			secrets[adminVaultPath] = map[string]interface{}{
				"password": adm.TmpPassword,
			}

			admins[i].PasswordVaultSecret = adminVaultPath
			admins[i].PasswordVaultSecretKey = "password"
			admins[i].TmpPassword = ""
		}

		values[AdministratorsValuesKey] = admins
		return true, nil
	}

	return false, nil
}

func (a *App) prepareDNSConfig(ginContext *gin.Context, r *registry, _values *Values,
	secretData map[string]map[string]interface{}, _ *[]string,
) (bool, error) {
	// TODO: add something to mrActions
	valuesChanged := false

	if r.DNSNameOfficerEnabled == "" && _values.Portals.Officer.CustomDNS.Enabled {
		_values.Portals.Officer.CustomDNS.Enabled = false
		valuesChanged = true
	} else if r.DNSNameOfficerEnabled != "" && r.DNSNameOfficer != "" {
		_values.Portals.Officer.CustomDNS = CustomDNS{Enabled: true, Host: r.DNSNameOfficer}
		valuesChanged = true

		certFile, _, err := ginContext.Request.FormFile("officer-ssl")
		if err == nil {
			certData, err := ioutil.ReadAll(certFile)
			if err != nil {
				return false, errors.Wrap(err, "unable to read officer ssl data")
			}

			pemInfo, err := DecodePEM(certData)
			if err != nil {
				return false, validator.ValidationErrors([]validator.FieldError{
					router.MakeFieldError("DNSNameOfficer", "pem-decode-error"),
				})
			}

			secretPath := strings.ReplaceAll(a.Config.VaultOfficerSSLPath, "{registry}", r.Name)
			secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameOfficer)

			if _, ok := secretData[secretPath]; !ok {
				secretData[secretPath] = make(map[string]interface{})
			}

			secretData[secretPath][VaultKeyCACert] = pemInfo.CACert
			secretData[secretPath][VaultKeyCert] = pemInfo.Cert
			secretData[secretPath][VaultKeyPK] = pemInfo.PrivateKey
		}
	}

	if r.DNSNameCitizenEnabled == "" && _values.Portals.Citizen.CustomDNS.Enabled {
		_values.Portals.Citizen.CustomDNS.Enabled = false
		valuesChanged = true
	} else if r.DNSNameCitizenEnabled != "" && r.DNSNameCitizen != "" {
		_values.Portals.Citizen.CustomDNS = CustomDNS{Host: r.DNSNameCitizen, Enabled: true}
		valuesChanged = true

		certFile, _, err := ginContext.Request.FormFile("citizen-ssl")
		if err == nil {
			certData, err := ioutil.ReadAll(certFile)
			if err != nil {
				return false, errors.Wrap(err, "unable to read citizen ssl data")
			}

			pemInfo, err := DecodePEM(certData)
			if err != nil {
				return false, validator.ValidationErrors([]validator.FieldError{
					router.MakeFieldError("DNSNameCitizen", "pem-decode-error"),
				})
			}

			secretPath := strings.ReplaceAll(a.Config.VaultCitizenSSLPath, "{registry}", r.Name)
			secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameCitizen)

			if _, ok := secretData[secretPath]; !ok {
				secretData[secretPath] = make(map[string]interface{})
			}

			secretData[secretPath][VaultKeyCACert] = pemInfo.CACert
			secretData[secretPath][VaultKeyCert] = pemInfo.Cert
			secretData[secretPath][VaultKeyPK] = pemInfo.PrivateKey
		}
	}

	if valuesChanged {
		_values.OriginalYaml[PortalsIndex] = _values.Portals
	}

	return valuesChanged, nil
}

func (a *App) isSmtpPasswordSet(path string) bool {
	data, err := a.Vault.Read(path)
	if err != nil && !errors.Is(err, vault.ErrSecretIsNil) {
		return false
	}
	password, ok := data[a.Config.VaultRegistrySMTPPwdSecretKey]
	if !ok {
		return false
	}
	if password == "" {
		return false
	}
	return true
}

func (a *App) prepareMailServerConfig(_ *gin.Context, r *registry, _values *Values,
	secretData map[string]map[string]interface{}, _ *[]string,
) (bool, error) {
	values := _values.OriginalYaml

	var email ExternalEmailSettings
	var vaultPath string

	if r.MailServerType == SMTPTypeExternal {
		var smptOptsDict map[string]string
		if err := json.Unmarshal([]byte(r.MailServerOpts), &smptOptsDict); err != nil {
			return false, errors.Wrap(err, "unable to decode mail server opts")
		}

		pwd, ok := smptOptsDict["password"]
		passwordExist := a.isSmtpPasswordSet(_values.Global.Notifications.Email.VaultPath)
		if !ok && !passwordExist {
			return false, errors.New("no password in mail server opts")
		}
		if pwd != "" {
			vaultPath = a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", "smtp", time.Now().Format("20060201T150405Z")))
			if _, ok := secretData[vaultPath]; !ok {
				secretData[vaultPath] = make(map[string]interface{})
			}

			secretData[vaultPath][a.Config.VaultRegistrySMTPPwdSecretKey] = pwd
		} else {
			vaultPath = _values.Global.Notifications.Email.VaultPath
		}

		// TODO: remove password from dict

		port, err := strconv.ParseInt(smptOptsDict["port"], 10, 32)
		if err != nil {
			return false, errors.Wrapf(err, "wrong smtp port value: %s", smptOptsDict["port"])
		}

		email = ExternalEmailSettings{
			Type:      "external",
			Host:      smptOptsDict["host"],
			Port:      port,
			Address:   smptOptsDict["address"],
			VaultPath: vaultPath,
			VaultKey:  a.Config.VaultRegistrySMTPPwdSecretKey,
		}
	} else {
		email = ExternalEmailSettings{
			Type: "internal",
		}
	}

	if reflect.DeepEqual(email, _values.Global.Notifications.Email) {
		return false, nil
	}

	globalInterface, ok := values[GlobalValuesIndex]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	globalDict["notifications"] = map[string]interface{}{
		"email": email,
	}
	values[GlobalValuesIndex] = globalDict

	return true, nil
}

func (a *App) prepareRegistryCodebase(r *registry) *codebase.Codebase {
	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
	framework := "helm"
	gitURL := codebase.RepoNotReady
	gitURLPath := "/" + r.Name
	cb := codebase.Codebase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1",
			Kind:       "Codebase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: codebase.CodebaseSpec{
			Description:      &r.Description,
			Type:             "registry",
			BuildTool:        "gitops",
			Lang:             "other",
			DefaultBranch:    r.RegistryGitBranch,
			EmptyProject:     false,
			Framework:        &framework,
			Strategy:         "import",
			DeploymentScript: "openshift-template",
			GitServer:        "gerrit",
			GitUrlPath:       &gitURLPath,
			CiTool:           "jenkins",
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				Url: gitURL,
			},
			JenkinsSlave: &jenkinsSlave,
		},
		Status: codebase.CodebaseStatus{
			Available:       false,
			LastTimeUpdated: time.Now(),
			Status:          "initialized",
			Action:          "codebase_registration",
			Value:           "inactive",
		},
	}

	if cb.Spec.DefaultBranch != "master" {
		cb.Spec.BranchToCopyInDefaultBranch = cb.Spec.DefaultBranch
		cb.Spec.DefaultBranch = "master"

		if a.EnableBranchProvisioners {
			jobProvisioning = branchProvisioner(cb.Spec.BranchToCopyInDefaultBranch)
			cb.Spec.JobProvisioning = &jobProvisioning
		}
	}

	if a.codebaseLabels != nil && len(a.codebaseLabels) > 0 {
		cb.SetLabels(a.codebaseLabels)
	}

	return &cb
}

func branchProvisioner(branch string) string {
	return "default-" + strings.Replace(
		strings.ToLower(branch), ".", "-", -1)
}
