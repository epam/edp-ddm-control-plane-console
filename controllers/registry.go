package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models"
	"ddm-admin-console/service"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
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

func (r *CreateRegistry) Post() {
	r.TplName = "registry/create.html"
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType

	var (
		err      error
		registry models.Registry
		valid    validation.Validation
	)
	defer func() {
		if err != nil {
			r.Ctx.ResponseWriter.WriteHeader(500)
			r.Data["error"] = err.Error()
		}
		if valid.HasErrors() {
			r.Ctx.ResponseWriter.WriteHeader(422)
			r.Data["errorsMap"] = valid.ErrorMap()
		}
	}()

	if err := r.ParseForm(&registry); err != nil {
		return
	}
	r.Data["model"] = registry

	b, err := valid.Valid(&registry)
	if err != nil || !b {
		return
	}

	if _, err := r.RegistryService.Create(registry.Name, registry.Description); err != nil {
		switch err.(type) {
		case service.RegistryExistsError:
			valid.AddError("Name.Required", err.Error())
			err = nil
			return
		default:
			return
		}
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
