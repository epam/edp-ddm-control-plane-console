package registry

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/k8s"

	"gopkg.in/yaml.v2"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationSMPTType        = "registry-parameters/smtp-type"
	AnnotationSMPTOpts        = "registry-parameters/smtp-opts"
	AnnotationTemplateName    = "registry-parameters/template-name"
	AnnotationCreatorUsername = "registry-parameters/creator-username"
	AnnotationCreatorEmail    = "registry-parameters/creator-email"
	AnnotationValues          = "registry-parameters/values"
)

func (a *App) createRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	prjs, err := a.Services.Gerrit.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	return router.MakeResponse(200, "registry/create.html", gin.H{
		"dnsManual":            false,
		"page":                 "registry",
		"gerritProjects":       prjs,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"hwINITemplateContent": hwINITemplateContent,
		"smtpConfig":           "{}",
	}), nil
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

func (a *App) filterProjects(projects []gerrit.GerritProject) []gerrit.GerritProject {
	filteredProjects := make([]gerrit.GerritProject, 0, 4)
	for _, prj := range projects {
		if strings.Contains(prj.Spec.Name, a.Config.GerritRegistryPrefix) {
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

func (a *App) getINITemplateContent() (string, error) {
	iniTemplate, err := os.Open(a.Config.HardwareINITemplatePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to open ini template file")
	}

	data, err := ioutil.ReadAll(iniTemplate)
	if err != nil {
		return "", errors.Wrap(err, "unable to read ini template data")
	}

	return string(data), nil
}

func (a *App) createRegistryPost(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

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

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	prjs, err := a.Services.Gerrit.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	r := registry{Scenario: ScenarioKeyRequired}
	if err := ctx.ShouldBind(&r); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/create.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "model": r,
				"hwINITemplateContent": hwINITemplateContent, "gerritProjects": prjs}), nil
	}

	if err := a.validateCreateRegistryGitTemplate(ctx, &r); err != nil {
		return nil, errors.Wrap(err, "unable to validate create registry git template")
	}

	if err := a.createRegistry(userCtx, ctx, &r, cbService, k8sService); err != nil {
		validationErrors, ok := errors.Cause(err).(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/create.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "model": r,
				"hwINITemplateContent": hwINITemplateContent, "gerritProjects": prjs}), nil
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

func (a *App) validateCreateRegistryGitTemplate(ctx context.Context, r *registry) error {
	prjs, err := a.Services.Gerrit.GetProjects(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	for _, prj := range prjs {
		if prj.Spec.Name == r.RegistryGitTemplate {
			for _, br := range prj.Status.Branches {
				if br == fmt.Sprintf("refs/heads/%s", r.RegistryGitBranch) {
					return nil
				}
			}
		}
	}

	return errors.New("wrong registry template selected")
}

func (a *App) createRegistry(ctx context.Context, ginContext *gin.Context, r *registry,
	cbService codebase.ServiceInterface, k8sService k8s.ServiceInterface) error {
	_, err := cbService.Get(r.Name)
	if err == nil {
		return validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Name", "registry-exists")})
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unknown error")
	}

	values, vaultSecretData := make(map[string]interface{}), make(map[string]map[string]interface{})
	if err := a.prepareDNSConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare dns config")
	}

	if err := a.prepareMailServerConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare mail server config")
	}

	if err := a.createVaultSecrets(vaultSecretData); err != nil {
		return errors.Wrap(err, "unable to create vault secrets")
	}

	if err := a.Services.Gerrit.CreateProject(ctx, r.Name); err != nil {
		return errors.Wrap(err, "unable to create gerrit project")
	}

	if r.Admins != "" {
		admins, err := validateAdmins(r.Admins)
		if err != nil {
			return errors.Wrap(err, "unable to validate admins")
		}

		//TODO: move admins creation to values yaml
		if err := a.admins.SyncAdmins(ctx, r.Name, admins); err != nil {
			return errors.Wrap(err, "unable to sync admins")
		}
	}

	if err := a.createRegistryKeys(r, ginContext.Request, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	cb := a.prepareRegistryCodebase(r)
	valuesEncoded, err := a.encodeValues(r.Name, values)
	if err != nil {
		return errors.Wrap(err, "unable to encode values")
	}

	cb.Annotations = map[string]string{
		AnnotationTemplateName:    r.RegistryGitTemplate,
		AnnotationSMPTType:        r.MailServerType, //TODO: remove
		AnnotationSMPTOpts:        r.MailServerOpts, //TODO: remove
		AnnotationCreatorUsername: ginContext.GetString(router.UserNameSessionKey),
		AnnotationCreatorEmail:    ginContext.GetString(router.UserEmailSessionKey),
		AnnotationValues:          valuesEncoded,
	}

	if err := cbService.Create(cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := cbService.CreateDefaultBranch(cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func (a *App) createVaultSecrets(secretData map[string]map[string]interface{}) error {
	for vaultPath, pathSecretData := range secretData {
		pathParts := strings.Split(vaultPath, "/")
		pathParts = append(pathParts[:1], append([]string{"data"}, pathParts[1:]...)...)
		vaultPath = strings.Join(pathParts, "/")

		if _, err := a.Vault.Write(vaultPath, map[string]interface{}{
			"data": pathSecretData,
		}); err != nil {
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

func (a *App) prepareDNSConfig(ginContext *gin.Context, r *registry, secretData map[string]map[string]interface{}, values map[string]interface{}) error {
	dns := make(map[string]string)

	if r.DNSNameOfficer != "" {
		dns["officerPortal"] = r.DNSNameOfficer

		certFile, _, err := ginContext.Request.FormFile("officer-ssl")
		if err != nil {
			return errors.Wrap(err, "unable to get officer ssl certificate")
		}

		certData, err := ioutil.ReadAll(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read officer ssl data")
		}

		caCert, cert, key, err := decodePEM(certData)
		if err != nil {
			return errors.Wrap(err, "unable to decode officer key file")
		}

		secretPath := strings.ReplaceAll(a.Config.VaultOfficerSSLPath, "{registry}", r.Name)
		secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameOfficer)

		if _, ok := secretData[secretPath]; !ok {
			secretData[secretPath] = make(map[string]interface{})
		}

		secretData[secretPath][a.Config.VaultOfficerCACertKey] = caCert
		secretData[secretPath][a.Config.VaultOfficerCertKey] = cert
		secretData[secretPath][a.Config.VaultOfficerPKKey] = key

		dns["officerPortalVaultCaKey"] = a.Config.VaultOfficerCACertKey
		dns["officerPortalVaultCertKey"] = a.Config.VaultOfficerCertKey
		dns["officerPortalVaultPKKey"] = a.Config.VaultOfficerPKKey
		dns["officerPortalVaultSecretPath"] = secretPath
	}

	if r.DNSNameCitizen != "" {
		dns["citizenPortal"] = r.DNSNameCitizen

		certFile, _, err := ginContext.Request.FormFile("citizen-ssl")
		if err != nil {
			return errors.Wrap(err, "unable to get citizen ssl certificate")
		}

		certData, err := ioutil.ReadAll(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read citizen ssl data")
		}

		caCert, cert, key, err := decodePEM(certData)
		if err != nil {
			return errors.Wrap(err, "unable to decode citizen key file")
		}

		secretPath := strings.ReplaceAll(a.Config.VaultCitizenSSLPath, "{registry}", r.Name)
		secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameCitizen)

		if _, ok := secretData[secretPath]; !ok {
			secretData[secretPath] = make(map[string]interface{})
		}

		secretData[secretPath][a.Config.VaultCitizenCACertKey] = caCert
		secretData[secretPath][a.Config.VaultCitizenCertKey] = cert
		secretData[secretPath][a.Config.VaultCitizenPKKey] = key

		dns["citizenPortalVaultCaKey"] = a.Config.VaultCitizenCACertKey
		dns["citizenPortalVaultCertKey"] = a.Config.VaultCitizenCertKey
		dns["citizenPortalVaultPKKey"] = a.Config.VaultCitizenPKKey
		dns["citizenPortalVaultSecretPath"] = secretPath
	}

	if len(dns) > 0 {
		values["customDNS"] = dns
	}

	return nil
}

func decodePEM(buf []byte) (caCert string, cert string, privateKey string, retErr error) {
	var (
		block   *pem.Block
		caBlock bytes.Buffer
	)

	for {
		var tmp bytes.Buffer

		block, buf = pem.Decode(buf)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			x509Cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				panic(err)
			}

			if x509Cert.IsCA {
				if retErr = pem.Encode(&caBlock, block); retErr != nil {
					return
				}
			} else {
				if retErr = pem.Encode(&tmp, block); retErr != nil {
					return
				}
				cert = tmp.String()
			}
		} else {
			if retErr = pem.Encode(&tmp, block); retErr != nil {
				return
			}

			privateKey = tmp.String()
		}
	}

	caCert = caBlock.String()

	if privateKey == "" {
		retErr = errors.New("no key found in PEM file")
	} else if caCert == "" {
		retErr = errors.New("no CA certs found in PEM file")
	} else if cert == "" {
		retErr = errors.New("no cert found in PEM file")
	}

	return
}

func (a *App) prepareMailServerConfig(ginContext *gin.Context, r *registry, secretData map[string]map[string]interface{}, values map[string]interface{}) error {
	action := ginContext.PostForm("action")
	if action == "edit" {
		performEdit := ginContext.PostForm("edit-smtp")
		if performEdit == "" {
			return nil
		}
	}

	notifications := make(map[string]interface{})

	if r.MailServerType == SMTPTypeExternal {
		var smptOptsDict map[string]string
		if err := json.Unmarshal([]byte(r.MailServerOpts), &smptOptsDict); err != nil {
			return errors.Wrap(err, "unable to decode mail server opts")
		}

		pwd, ok := smptOptsDict["password"]
		if !ok {
			return errors.New("no password in mail server opts")
		}

		if _, ok := secretData[a.vaultRegistryPath(r.Name)]; !ok {
			secretData[a.vaultRegistryPath(r.Name)] = make(map[string]interface{})
		}

		secretData[a.vaultRegistryPath(r.Name)][a.Config.VaultRegistrySMTPPwdSecretKey] = pwd
		//TODO: remove password from dict

		port, err := strconv.ParseInt(smptOptsDict["port"], 10, 32)
		if err != nil {
			return errors.Wrapf(err, "wrong smtp port value: %s", smptOptsDict["port"])
		}

		notifications["email"] = map[string]interface{}{
			"type":      "external",
			"host":      smptOptsDict["host"],
			"port":      port,
			"address":   smptOptsDict["address"],
			"password":  smptOptsDict["password"],
			"vaultPath": a.vaultRegistryPath(r.Name),
			"vaultKey":  a.Config.VaultRegistrySMTPPwdSecretKey,
		}
	} else {
		notifications["email"] = map[string]interface{}{
			"type": "internal",
		}
	}

	values["notifications"] = notifications

	return nil
}

func (a *App) prepareRegistryCodebase(r *registry) *codebase.Codebase {
	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
	gitURL := codebase.RepoNotReady
	cb := codebase.Codebase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
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
			Strategy:         "import",
			DeploymentScript: "openshift-template",
			GitServer:        "gerrit",
			GitUrlPath:       &gitURL,
			CiTool:           "Jenkins",
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				//Url: fmt.Sprintf("%s/%s", gerritRegistryHost, r.RegistryGitTemplate),
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

func validateRegistryKeys(rq *http.Request, r *registry) (createKeys bool, key6Fl, caCertFl,
	caJSONFl multipart.File, err error) {

	var fieldErrors []validator.FieldError
	caCertFl, _, err = rq.FormFile("ca-cert")
	if err != nil {
		if !r.KeysRequired() {
			err = nil
			return
		}

		fieldErrors = append(fieldErrors, router.MakeFieldError("CACertificate", "required"))
	}

	caJSONFl, _, err = rq.FormFile("ca-json")
	if err != nil {
		fieldErrors = append(fieldErrors, router.MakeFieldError("CAsJSON", "required"))
	}

	if r.KeyDeviceType == KeyDeviceTypeFile {
		key6Fl, _, err = rq.FormFile("key6")
		if err != nil {
			fieldErrors = append(fieldErrors, router.MakeFieldError("Key6", "required"))
		}
	}

	if len(fieldErrors) > 0 {
		err = validator.ValidationErrors(fieldErrors)
		return
	}

	createKeys = true
	return
}

func (a *App) createRegistryKeys(reg *registry, rq *http.Request, k8sService k8s.ServiceInterface) error {
	createKeys, key6Fl, caCertFl, caJSONFl, err := validateRegistryKeys(rq, reg)
	if err != nil {
		return errors.Wrap(err, "unable to validate registry keys")
	}
	if !createKeys {
		return nil
	}

	filesSecretData := make(map[string][]byte)
	envVarsSecretData := map[string][]byte{
		"sign.key.device-type": []byte(reg.KeyDeviceType),
	}

	if err := a.setCASecretData(filesSecretData, caCertFl, caJSONFl); err != nil {
		return errors.Wrap(err, "unable to set ca secret data for registry")
	}

	if err := a.setKeySecretDataFromRegistry(reg, key6Fl, filesSecretData, envVarsSecretData); err != nil {
		return errors.Wrap(err, "unable to set key vars from registry form")
	}

	if err := a.setAllowedKeysSecretData(filesSecretData, reg); err != nil {
		return errors.Wrap(err, "unable to set allowed keys secret data")
	}

	if err := k8sService.RecreateSecret(fmt.Sprintf("digital-signature-ops-%s-data", reg.Name),
		filesSecretData); err != nil {
		return errors.Wrap(err, "unable to create secret")
	}

	if err := k8sService.RecreateSecret(fmt.Sprintf("digital-signature-ops-%s-env-vars", reg.Name),
		envVarsSecretData); err != nil {
		return errors.Wrap(err, "unable to create secret")
	}

	return nil
}

func (a *App) setKeySecretDataFromRegistry(reg *registry, key6Fl multipart.File,
	filesSecretData, envVarsSecretData map[string][]byte) error {

	if reg.KeyDeviceType == KeyDeviceTypeFile {
		key6Bytes, err := ioutil.ReadAll(key6Fl)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}
		filesSecretData["Key-6.dat"] = key6Bytes
		envVarsSecretData["sign.key.file.issuer"] = []byte(reg.SignKeyIssuer)
		envVarsSecretData["sign.key.file.password"] = []byte(reg.SignKeyPwd)

		//TODO: temporary hack, remote in future
		envVarsSecretData["sign.key.hardware.type"] = []byte{}
		envVarsSecretData["sign.key.hardware.device"] = []byte{}
		envVarsSecretData["sign.key.hardware.password"] = []byte{}
		filesSecretData["osplm.ini"] = []byte{}
		// end todo

	} else if reg.KeyDeviceType == KeyDeviceTypeHardware {
		envVarsSecretData["sign.key.hardware.type"] = []byte(reg.RemoteType)
		envVarsSecretData["sign.key.hardware.device"] = []byte(fmt.Sprintf("%s:%s (%s)",
			reg.RemoteSerialNumber, reg.RemoteKeyPort, reg.RemoteKeyHost))
		envVarsSecretData["sign.key.hardware.password"] = []byte(reg.RemoteKeyPassword)
		filesSecretData["osplm.ini"] = []byte(reg.INIConfig)

		//TODO: temporary hack, remote in future
		filesSecretData["Key-6.dat"] = []byte{}
		envVarsSecretData["sign.key.file.issuer"] = []byte{}
		envVarsSecretData["sign.key.file.password"] = []byte{}
		// end todo
	}

	return nil
}

func (a *App) setCASecretData(filesSecretData map[string][]byte, caCertFl, caJSONFl multipart.File) error {
	caCertBytes, err := ioutil.ReadAll(caCertFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}
	filesSecretData["CACertificates.p7b"] = caCertBytes

	casJSONBytes, err := ioutil.ReadAll(caJSONFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}
	filesSecretData["CAs.json"] = casJSONBytes

	return nil
}

func (a *App) setAllowedKeysSecretData(filesSecretData map[string][]byte, reg *registry) error {
	//TODO tmp hack, remote in future
	filesSecretData["allowed-keys.yml"] = []byte{}
	//end todo

	if len(reg.AllowedKeysIssuer) > 0 {
		var allowedKeysConf allowedKeysConfig
		for i := range reg.AllowedKeysIssuer {
			allowedKeysConf.AllowedKeys = append(allowedKeysConf.AllowedKeys, allowedKey{
				Issuer: reg.AllowedKeysIssuer[i],
				Serial: reg.AllowedKeysSerial[i],
			})
		}
		allowedKeysYaml, err := yaml.Marshal(&allowedKeysConf)
		if err != nil {
			return errors.Wrap(err, "unable to encode allowed keys to yaml")
		}
		filesSecretData["allowed-keys.yml"] = allowedKeysYaml
	}

	return nil
}
