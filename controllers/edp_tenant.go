/*
 * Copyright 2020 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/service"
	ec "ddm-admin-console/service/edp-component"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
)

type EDPTenantController struct {
	beego.Controller
	EDPTenantService service.EDPTenantService
	EDPComponent     ec.Service
}

func (c *EDPTenantController) GetEDPComponents() {
	comp, err := c.EDPComponent.GetEDPComponents()
	if err != nil {
		c.Abort("500")
		return
	}

	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["InputURL"] = strings.TrimSuffix(c.Ctx.Input.URL(), "/"+console.Tenant)
	c.Data["EDPTenantName"] = console.Tenant
	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["EDPComponents"] = comp
	c.Data["Type"] = "overview"
	c.Data["BasePath"] = console.BasePath
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.TplName = "edp_components.html"
}

func (c *EDPTenantController) GetVcsIntegrationValue() {
	isVcsEnabled, err := c.EDPTenantService.GetVcsIntegrationValue()

	if err != nil {
		if err.Error() == "NOT_FOUND" {
			http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Data["json"] = isVcsEnabled
	c.ServeJSON()
}
