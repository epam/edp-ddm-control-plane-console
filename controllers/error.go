package controllers

import (
	"github.com/astaxie/beego"
)

const (
	notFoundTemplatePath = "error/error_404.html"
)

type ErrorController struct {
	beego.Controller
	BasePath string
}

func MakeErrorController(basePath string) *ErrorController {
	return &ErrorController{
		BasePath: basePath,
	}
}

func (t *ErrorController) Error500() {
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = t.BasePath
	t.TplName = "error/error_500.html"
}

func (t *ErrorController) Error403() {
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = t.BasePath
	t.TplName = "error/error_403.html"
}

func (t *ErrorController) Error404() {
	t.Data["Username"] = t.Ctx.Input.Session("username")
	t.Data["BasePath"] = t.BasePath
	t.TplName = notFoundTemplatePath
}
