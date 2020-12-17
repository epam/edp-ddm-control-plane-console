package controllers

import (
	"github.com/astaxie/beego"
)

const registryType = "registry"

type ListRegistry struct {
	beego.Controller
}

func (r *ListRegistry) Get() {
	r.Data["Type"] = registryType
	r.TplName = "registry/list.html"
}

type CreateRegistry struct {
	beego.Controller
}

func (r *CreateRegistry) Get() {
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/create.html"
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
