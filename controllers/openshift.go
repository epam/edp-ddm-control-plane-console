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
	"net/http"
)

type OpenshiftRestController struct {
	beego.Controller
	ClusterService service.ClusterService
}

func (c *OpenshiftRestController) GetAllStorageClasses() {
	storageClasses, err := c.ClusterService.GetAllStorageClasses()

	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Data["json"] = storageClasses
	c.ServeJSON()
}