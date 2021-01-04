package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/pkg/errors"
)

const (
	registryType     = "registry"
	defaultBranch    = "master"
	lang             = "other"
	strategy         = "clone"
	deploymentScript = "openshift-template"
	ciTool           = "Jenkins"
)

type CodebaseService interface {
	CreateCodebase(codebase command.CreateCodebase) (*edpv1alpha1.Codebase, error)
	GetCodebasesByCriteria(criteria query.CodebaseCriteria) ([]*query.Codebase, error)
	GetCodebaseByName(name string) (*query.Codebase, error)
	UpdateDescription(name, description string) error
}

type ListRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
}

func MakeListRegistry(codebaseService CodebaseService) *ListRegistry {
	return &ListRegistry{
		CodebaseService: codebaseService,
	}
}

func (r *ListRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["Type"] = registryType

	r.TplName = "registry/list.html"

	codebases, err := r.CodebaseService.GetCodebasesByCriteria(query.CodebaseCriteria{
		Type: query.Registry,
	})

	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Data["registries"] = codebases
}

type EditRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
}

func MakeEditRegistry(codebaseService CodebaseService) *EditRegistry {
	return &EditRegistry{
		CodebaseService: codebaseService,
	}
}

func (r *EditRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	registryName := r.Ctx.Input.Param(":name")
	rg, err := r.CodebaseService.GetCodebaseByName(registryName)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Data["description"] = rg.Description
}

func (r *EditRegistry) editRegistry(registry *models.Registry) (errorMap map[string][]*validation.Error,
	err error) {
	var valid validation.Validation

	dataValid, err := valid.Valid(registry)
	if err != nil {
		return nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
		return valid.ErrorMap(), nil
	}

	if err := r.CodebaseService.UpdateDescription(registry.Name, registry.Description); err != nil {
		return nil, errors.Wrap(err, "something went wrong during k8s registry edit")
	}

	return
}

func (r *EditRegistry) Post() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	var parsedRegistry models.Registry
	registryName := r.Ctx.Input.Param(":name")
	if err := r.ParseForm(&parsedRegistry); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}
	parsedRegistry.Name = registryName

	validationErrors, err := r.editRegistry(&parsedRegistry)
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

	r.Redirect("/admin/registry/overview", 303)
}

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

	_, err = r.CodebaseService.CreateCodebase(command.CreateCodebase{
		Name:             registry.Name,
		Username:         username,
		Type:             string(query.Registry),
		Description:      &registry.Description,
		DefaultBranch:    defaultBranch,
		Lang:             lang,
		BuildTool:        lang,
		Strategy:         strategy,
		DeploymentScript: deploymentScript,
		GitServer:        defaultGitServer,
		CiTool:           ciTool,
		Repository: &command.Repository{
			URL: beego.AppConfig.String("defaultGitRepo"),
		},
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

	r.Redirect("/admin/registry/overview", 303)
}

type ViewRegistry struct {
	beego.Controller
}

func (r *ViewRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["Type"] = registryType
	r.TplName = "registry/view.html"
}
