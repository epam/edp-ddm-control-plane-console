package controllers

import (
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/util"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
)

const (
	key6FormKey      = "key6"
	caCertFormKey    = "ca-cert"
	caJSONFormKey    = "ca-json"
	sessionGroupsKey = "groups"
)

type CreateRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
	BasePath        string
	CreatorGroup    string
	GroupValidator  GroupValidator
}

func MakeCreateRegistry(basePath, creatorGroup string, codebaseService CodebaseService) *CreateRegistry {
	return &CreateRegistry{
		CodebaseService: codebaseService,
		BasePath:        basePath,
		CreatorGroup:    creatorGroup,
		GroupValidator:  &groupValidator{},
	}
}

func (r *CreateRegistry) Get() {
	r.Data["BasePath"] = r.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/create.html"
}

func (r *CreateRegistry) Post() {
	r.Data["BasePath"] = r.BasePath
	r.TplName = "registry/create.html"
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType

	var registry models.Registry
	if err := r.ParseForm(&registry); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
		return
	}
	r.Data["model"] = registry

	validationErrors, err := r.createRegistry(&registry, r.Ctx.Request)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
		return
	}

	if validationErrors != nil {
		log.Error(fmt.Sprintf("%+v\n", validationErrors))
		r.Data["errorsMap"] = validationErrors
		r.Ctx.Output.Status = 422
		if err := r.Render(); err != nil {
			log.Error(err.Error())
		}
		return
	}

	r.Redirect(registryOverviewURL, 303)
}

func (r *CreateRegistry) createRegistry(registry *models.Registry,
	request *http.Request) (errorMap map[string][]*validation.Error, err error) {
	if !r.GroupValidator.IsAllowedToCreateRegistry(r.Ctx, r.CreatorGroup) {
		return nil, errors.New("user is not allowed to create registry")
	}

	var valid validation.Validation

	dataValid, err := valid.Valid(registry)
	if err != nil {
		return nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
		return valid.ErrorMap(), nil
	}

	if err = createRegistryKeys(r.CodebaseService, registry, &valid, request, true); err != nil {
		err = errors.Wrap(err, "unable to create registry keys")
	}

	if len(valid.ErrorMap()) > 0 {
		return valid.ErrorMap(), nil
	}

	username, _ := r.Ctx.Input.Session("username").(string)
	jobProvisioning := "default"
	startVersion := "0.0.1"
	versioning := command.Versioning{
		StartFrom: &startVersion,
		Type:      "edp",
	}

	_, err = r.CodebaseService.CreateCodebase(command.CreateCodebase{
		Name:             registry.Name,
		Username:         username,
		Type:             string(query.Registry),
		Description:      &registry.Description,
		DefaultBranch:    defaultBranch,
		Lang:             lang,
		BuildTool:        buildTool,
		Strategy:         strategy,
		DeploymentScript: deploymentScript,
		GitServer:        defaultGitServer,
		CiTool:           ciTool,
		JobProvisioning:  &jobProvisioning,
		Versioning:       versioning,
		Repository: &command.Repository{
			URL: beego.AppConfig.String("registryGitRepo"),
		},
		JenkinsSlave: util.GetStringP(jenkinsSlave),
		Admins:       registry.Admins,
	})

	if err != nil {
		switch err.(type) {
		case *edperror.CodebaseAlreadyExistsError:
			valid.AddError("Name.Required", err.Error())
			return valid.ErrorMap(), nil
		default:
			return nil, errors.Wrap(err, "something went wrong during codebase creation")
		}
	}

	return
}

func validateRegistryKeys(registry *models.Registry, valid *validation.Validation,
	rq *http.Request, required bool) (createKeys bool, key6Fl, caCertFl, caJSONFl multipart.File) {

	var err error
	key6Fl, _, err = rq.FormFile(key6FormKey)
	if err != nil {
		if !required {
			err = nil
			return
		}

		valid.AddError("Key6.Required", err.Error())
	}

	caCertFl, _, err = rq.FormFile(caCertFormKey)
	if err != nil {
		valid.AddError("CACertificate.Required", err.Error())
	}

	caJSONFl, _, err = rq.FormFile(caJSONFormKey)
	if err != nil {
		valid.AddError("CAsJSON.Required", err.Error())
	}

	if registry.SignKeyPwd == "" {
		valid.AddError("SignKeyPwd.Required", "Can not be empty")
	}

	if registry.SignKeyIssuer == "" {
		valid.AddError("SignKeyIssuer.Required", "Can not be empty")
	}

	createKeys = true

	return
}

func createRegistryKeys(cb CodebaseService, registry *models.Registry, valid *validation.Validation, rq *http.Request,
	required bool) error {

	createKeys, key6Fl, caCertFl, caJSONFl := validateRegistryKeys(registry, valid, rq, required)
	if !createKeys || len(valid.ErrorsMap) > 0 {
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

	if err := cb.CreateKeySecret(key6Bytes, caCertBytes, casJSONBytes, registry.SignKeyIssuer,
		registry.SignKeyPwd, registry.Name); err != nil {
		return errors.Wrap(err, "unable to create registry keys secret")
	}

	return nil
}
