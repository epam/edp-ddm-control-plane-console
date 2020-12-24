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

type RegistryService interface {
	List() ([]*models.Registry, error)
	Create(name, description string) (*models.Registry, error)
	Get(name string) (*models.Registry, error)
	EditDescription(name, description string) error
}

type ListRegistry struct {
	beego.Controller
	RegistryService RegistryService
}

func MakeListRegistry(registryService RegistryService) *ListRegistry {
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
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
		return
	}

	r.Data["registries"] = registries
}

type CreateRegistry struct {
	beego.Controller
	RegistryService RegistryService
}

func MakeCreateRegistry(registryService RegistryService) *CreateRegistry {
	return &CreateRegistry{
		RegistryService: registryService,
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

	r.Redirect("/admin/edp/registry/overview", 303)
}

type EditRegistry struct {
	beego.Controller
	RegistryService RegistryService
}

func MakeEditRegistry(registryService RegistryService) *EditRegistry {
	return &EditRegistry{
		RegistryService: registryService,
	}
}

func (r *EditRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	registryName := r.Ctx.Input.Param(":name")
	rg, err := r.RegistryService.Get(registryName)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
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

	if err := r.RegistryService.EditDescription(registry.Name, registry.Description); err != nil {
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
		r.CustomAbort(500, err.Error())
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

	r.Redirect("/admin/edp/registry/overview", 303)
}

type ViewRegistry struct {
	beego.Controller
}

func (r *ViewRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["Type"] = registryType
	r.TplName = "registry/view.html"
}
