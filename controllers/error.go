package controllers

import (
	"ddm-admin-console/console"
	"github.com/astaxie/beego"
)

type ErrorController struct {
	beego.Controller
}

func (t *ErrorController) Error500() {
	t.Data["EDPVersion"] = console.EDPVersion
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = console.BasePath
	t.TplName = "error/error_500.html"
}

func (t *ErrorController) Error403() {
	t.Data["EDPVersion"] = console.EDPVersion
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = console.BasePath
	t.TplName = "error/error_403.html"
}

func (t *ErrorController) Error404() {
	t.Data["EDPVersion"] = console.EDPVersion
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = console.BasePath
	t.TplName = "error/error_404.html"
}
