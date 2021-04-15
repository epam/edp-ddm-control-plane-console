package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/util"
	"encoding/base64"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
)

type CreateRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
}

func MakeCreateRegistry(codebaseService CodebaseService) *CreateRegistry {
	return &CreateRegistry{
		CodebaseService: codebaseService,
	}
}

func (r *CreateRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/create.html"
}

func (r *CreateRegistry) Post() {
	r.Data["BasePath"] = console.BasePath
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

	validationErrors, err := r.createRegistry(&registry)
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

func (r *CreateRegistry) createRegistry(registry *models.Registry) (errorMap map[string][]*validation.Error,
	err error) {
	var valid validation.Validation

	dataValid, err := valid.Valid(registry)
	if err != nil {
		return nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
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

	if err := r.createRegistryKeys(registry, &valid); err != nil {
		err = errors.Wrap(err, "unable to create registry keys")
	}

	return
}

func (r *CreateRegistry) createRegistryKeys(registry *models.Registry, valid *validation.Validation) error {
	if registry.Key6 == "" {
		return nil
	}

	key6, err := base64.StdEncoding.DecodeString(registry.Key6)
	if err != nil {
		return errors.Wrapf(err, "unable to decode key6 base64")
	}

	if registry.SignKeyIssuer == "" {
		valid.AddError("SignKeyIssuer.Required", "digital-signature-key-issuer is required")
	}

	if registry.SignKeyPwd == "" {
		valid.AddError("SignKeyPwd.Required", "digital-signature-key-password is required")
	}

	if registry.CACertificate == "" {
		valid.AddError("CACertificate.Required", "CACertificates.p7b is required")
	}

	if registry.CAsJSON == "" {
		valid.AddError("CAsJSON.Required", "CAs.json is required")
	}

	caCert, err := base64.StdEncoding.DecodeString(registry.CACertificate)
	if err != nil {
		return errors.Wrapf(err, "unable to decode CACertificates.p7b base64")
	}

	casJSON, err := base64.StdEncoding.DecodeString(registry.CAsJSON)
	if err != nil {
		return errors.Wrapf(err, "unable to decode CAs.json base64")
	}

	if err := r.CodebaseService.CreateKeySecret(key6, caCert, casJSON, registry.SignKeyIssuer, registry.SignKeyPwd,
		registry.Name); err != nil {
		return errors.Wrap(err, "unable to create registry keys secret")
	}

	return nil
}
