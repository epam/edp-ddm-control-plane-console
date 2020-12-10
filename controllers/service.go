package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/service"
	"github.com/astaxie/beego"
)

type ThirdPartyServiceController struct {
	beego.Controller
	ThirdPartyService service.ThirdPartyService
}

func (s *ThirdPartyServiceController) GetServicePage() {
	services, err := s.ThirdPartyService.GetAllServices()
	if err != nil {
		s.Abort("500")
		return
	}

	s.Data["EDPVersion"] = console.EDPVersion
	s.Data["Username"] = s.Ctx.Input.Session("username")
	s.Data["Services"] = services
	s.Data["Type"] = "services"
	s.Data["BasePath"] = console.BasePath
	s.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	s.TplName = "service.html"
}
