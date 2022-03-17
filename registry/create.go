package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
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
	AdminsAnnotation       = "registry-parameters/administrators"
	GroupAnnotation        = "registry-parameters/group"
	TemplateNameAnnotation = "registry-parameters/template-name"
)

func (a *App) createRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	prjs, err := a.gerritService.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	groups, err := a.codebaseService.GetAllByType(codebase.GroupCodebaseType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get groups")
	}

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	return router.MakeResponse(200, "registry/create.html", gin.H{
		"page":                 "registry",
		"gerritProjects":       prjs,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"hwINITemplateContent": hwINITemplateContent,
		"groups":               groups,
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
		if strings.Contains(prj.Spec.Name, a.gerritRegistryPrefix) {
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
	iniTemplate, err := os.Open(a.hardwareINITemplatePath)
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

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get k8s service for user")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	hwINITemplateContent, err := a.getINITemplateContent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	prjs, err := a.gerritService.GetProjects(context.Background())
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

	if err := a.validateCreateRegistryGitTemplate(&r); err != nil {
		return nil, errors.Wrap(err, "unable to validate create registry git template")
	}

	if err := a.createRegistry(&r, ctx.Request, cbService, k8sService); err != nil {
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
	allowedToCreate, err := a.codebaseService.CheckIsAllowedToCreate(userK8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check create access")
	}
	if !allowedToCreate {
		return errors.New("access denied")
	}

	return nil
}

func (a *App) validateCreateRegistryGitTemplate(r *registry) error {
	prjs, err := a.gerritService.GetProjects(context.Background())
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

func (a *App) createRegistry(r *registry, request *http.Request, cbService codebase.ServiceInterface,
	k8sService k8s.ServiceInterface) error {
	_, err := cbService.Get(r.Name)
	if err == nil {
		return validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Name", "registry-exists")})
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unknown error")
	}

	if err := a.createRegistryKeys(r, request, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	cb := prepareRegistryCodebase(a.gerritRegistryHost, r)

	annotations := map[string]string{
		TemplateNameAnnotation: r.RegistryGitTemplate,
	}

	if r.Admins != "" {
		if err := validateAdmins(r.Admins); err != nil {
			return err
		}

		annotations[AdminsAnnotation] = base64.StdEncoding.EncodeToString([]byte(r.Admins))
	}
	if r.RegistryGroup != "" {
		annotations[GroupAnnotation] = base64.StdEncoding.EncodeToString([]byte(r.RegistryGroup))
	}
	cb.Annotations = annotations

	if err := cbService.Create(cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := cbService.CreateTempSecrets(cb, k8sService, a.gerritCreatorSecretName); err != nil {
		return errors.Wrap(err, "unable to create temp secrets")
	}

	if err := cbService.CreateDefaultBranch(cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func prepareRegistryCodebase(gerritRegistryHost string, r *registry) *codebase.Codebase {
	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
	return &codebase.Codebase{
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
			Strategy:         "clone",
			DeploymentScript: "openshift-template",
			GitServer:        "gerrit",
			CiTool:           "Jenkins",
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				Url: fmt.Sprintf("%s/%s", gerritRegistryHost, r.RegistryGitTemplate),
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
