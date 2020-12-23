package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models"
	"ddm-admin-console/service"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
)

const registryType = "registry"

type ListRegistry struct {
	beego.Controller
	RegistryService *service.Registry
}

func MakeListRegistry(registryService *service.Registry) *ListRegistry {
	return &ListRegistry{
		RegistryService: registryService,
	}
}

func (r *ListRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["Type"] = registryType

	r.TplName = "registry/list.html"

	registries, err := r.RegistryService.List()
	if err != nil {
		r.Data["error"] = err.Error()
		r.Ctx.ResponseWriter.WriteHeader(500)
		return
	}

	r.Data["registries"] = registries
}

type CreateRegistry struct {
	beego.Controller
	RegistryService *service.Registry
}

func MakeCreateRegistry(registryService *service.Registry) *CreateRegistry {
	return &CreateRegistry{
		RegistryService: registryService,
	}
}

func (r *CreateRegistry) Get() {
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

	if _, err := r.RegistryService.Create(registry.Name, registry.Description); err != nil {
		switch err.(type) {
		case service.RegistryExistsError:
			valid.AddError("Name.Required", err.Error())
			return valid.ErrorMap(), nil
		default:
			return nil, errors.Wrap(err, "something went wrong during registry creation")
		}
	}

	return
}

func (r *CreateRegistry) Post() {
	r.TplName = "registry/create.html"
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType

	var registry models.Registry
	if err := r.ParseForm(&registry); err != nil {
		return
	}
	r.Data["model"] = registry

	validationErrors, err := r.createRegistry(&registry)
	if err != nil {
		r.Ctx.ResponseWriter.WriteHeader(500)
		r.Data["error"] = err.Error()
		log.Error(fmt.Sprintf("%+v\n", err))
	}

	if validationErrors != nil {
		r.Ctx.ResponseWriter.WriteHeader(422)
		r.Data["errorsMap"] = validationErrors
	}

	r.Redirect("/admin/edp/registry/overview", 303)
}

type EditRegistry struct {
	beego.Controller
}

func (r *EditRegistry) Get() {
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"
}

type ViewRegistry struct {
	beego.Controller
}

func (r *ViewRegistry) Get() {
	r.Data["Type"] = registryType
	r.TplName = "registry/view.html"
}
