package controllers

import "github.com/astaxie/beego"

type ListRegistry struct {
	beego.Controller
}

func (r *ListRegistry) Get() {
	r.Data["Type"] = "registry"
	r.TplName = "registry/list.html"
}
