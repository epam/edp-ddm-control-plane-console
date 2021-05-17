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
	"ddm-admin-console/service"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
	EDPTenantService *service.EDPTenantService
	BasePath         string
}

func MakeMainController(basePath string, edpTenantService *service.EDPTenantService) *MainController {
	return &MainController{
		BasePath:         basePath,
		EDPTenantService: edpTenantService,
	}
}

func (c *MainController) Index() {
	c.Data["BasePath"] = c.BasePath
	c.TplName = "index.html"
}

type DashboardController struct {
	beego.Controller
}

func (c *DashboardController) Get() {
	c.Data["GerritLink"] = beego.AppConfig.String("gerritGlobalLink")
	c.Data["JenkinsLink"] = beego.AppConfig.String("jenkinsGlobalLink")
	c.Data["Type"] = "dashboard"
	c.TplName = "dashboard.html"
}

func MakeDashboardController() *DashboardController {
	return &DashboardController{}
}
