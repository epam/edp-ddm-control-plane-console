package controllers

import "github.com/astaxie/beego"

type RegistryController struct {
	beego.Controller
}

func (r *RegistryController) Get() {
	r.Data["Type"] = "registry"
	r.TplName = "registry/list.html"
}
