package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	key6FormKey                          = "key6"
	caCertFormKey                        = "ca-cert"
	caJSONFormKey                        = "ca-json"
	key6SecretKey                        = "Key-6.dat"
	digitalSignatureKeyIssuerSecretKey   = "digital-signature-key-issuer"
	digitalSignatureKeyPasswordSecretKey = "digital-signature-key-password"
	caCertificatesSecretKey              = "CACertificates.p7b"
	CAsJSONSecretKey                     = "CAs.json"
	AdminsAnnotation                     = "registry-parameters/administrators"
	gerritCreatorUsername                = "user"
	gerritCreatorPassword                = "password"
)

func (a *App) createRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	k8sService, err := a.k8sService.ServiceForContext(a.router.ContextWithUserAccessToken(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "create", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	if !allowedToCreate {
		return nil, errors.Wrap(err, "access denied")
	}

	return router.MakeResponse(200, "registry/create.html", gin.H{
		"page": "registry",
	}), nil
}

func (a *App) createRegistryPost(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "create", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	if !allowedToCreate {
		return nil, errors.Wrap(err, "access denied")
	}

	var r registry
	if err := ctx.ShouldBind(&r); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/create.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "model": r}), nil
	}

	if err := a.createRegistry(&r, ctx.Request, cbService, k8sService); err != nil {
		validationErrors, ok := errors.Cause(err).(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/create.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "model": r}), nil
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
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

	if err := a.createRegistryKeys(r, request, true, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	//username, _ := r.Ctx.Input.Session("username").(string)
	//TODO: get username from session

	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
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
			DefaultBranch:    "master",
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
				Url: a.registryGitRepo,
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

	annotations := make(map[string]string)
	if r.Admins != "" {
		annotations[AdminsAnnotation] = base64.StdEncoding.EncodeToString([]byte(r.Admins))
	}
	cb.Annotations = annotations

	if err := cbService.Create(&cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := a.createTempSecrets(&cb, k8sService); err != nil {
		return errors.Wrap(err, "unable to create temp secrets")
	}

	if err := cbService.CreateDefaultBranch(&cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func validateRegistryKeys(registry *registry, rq *http.Request, required bool) (createKeys bool, key6Fl, caCertFl,
	caJSONFl multipart.File, err error) {

	var fieldErrors []validator.FieldError
	key6Fl, _, err = rq.FormFile(key6FormKey)
	if err != nil {
		if !required {
			err = nil
			return
		}

		fieldErrors = append(fieldErrors, router.MakeFieldError("Key6", "required"))
	}

	caCertFl, _, err = rq.FormFile(caCertFormKey)
	if err != nil {
		fieldErrors = append(fieldErrors, router.MakeFieldError("CACertificate", "required"))
	}

	caJSONFl, _, err = rq.FormFile(caJSONFormKey)
	if err != nil {
		fieldErrors = append(fieldErrors, router.MakeFieldError("CAsJSON", "required"))
	}

	if registry.SignKeyPwd == "" {
		fieldErrors = append(fieldErrors, router.MakeFieldError("SignKeyPwd", "required"))
	}

	if registry.SignKeyIssuer == "" {
		fieldErrors = append(fieldErrors, router.MakeFieldError("SignKeyIssuer", "required"))
	}

	if len(fieldErrors) > 0 {
		err = validator.ValidationErrors(fieldErrors)
		return
	}

	createKeys = true
	return
}

func (a *App) createRegistryKeys(registry *registry, rq *http.Request, required bool, k8sService k8s.ServiceInterface) error {
	createKeys, key6Fl, caCertFl, caJSONFl, err := validateRegistryKeys(registry, rq, required)
	if err != nil {
		return errors.Wrap(err, "unable to validate registry keys")
	}
	if !createKeys {
		return nil
	}

	key6Bytes, err := ioutil.ReadAll(key6Fl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}

	caCertBytes, err := ioutil.ReadAll(caCertFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}

	casJSONBytes, err := ioutil.ReadAll(caJSONFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}

	if err := k8sService.RecreateSecret(fmt.Sprintf("system-digital-sign-%s-key", registry.Name),
		map[string][]byte{
			key6SecretKey:                        key6Bytes,
			digitalSignatureKeyIssuerSecretKey:   []byte(registry.SignKeyIssuer),
			digitalSignatureKeyPasswordSecretKey: []byte(registry.SignKeyPwd),
		}); err != nil {
		return errors.Wrap(err, "unable to create secret")
	}

	if err := k8sService.RecreateSecret(fmt.Sprintf("system-digital-sign-%s-ca", registry.Name),
		map[string][]byte{
			caCertificatesSecretKey: caCertBytes,
			CAsJSONSecretKey:        casJSONBytes,
		}); err != nil {
		return errors.Wrap(err, "unable to create secret")
	}

	return nil
}

func (a *App) createTempSecrets(cb *codebase.Codebase, k8sService k8s.ServiceInterface) error {
	secret, err := k8sService.GetSecret(a.gerritCreatorSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to get secret")
	}

	username, ok := secret.Data[gerritCreatorUsername]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have username")
	}

	pwd, ok := secret.Data[gerritCreatorPassword]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have password")
	}

	repoSecretName := fmt.Sprintf("repository-codebase-%s-temp", cb.Name)
	if err := k8sService.RecreateSecret(repoSecretName, map[string][]byte{
		"username": username,
		"password": pwd,
	}); err != nil {
		return errors.Wrapf(err, "unable to create secret: %s", repoSecretName)
	}

	return nil
}
