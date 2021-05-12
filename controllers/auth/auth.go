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

package auth

import (
	"context"
	"ddm-admin-console/auth"
	"ddm-admin-console/service/logger"
	"fmt"

	"github.com/astaxie/beego"
)

var log = logger.GetLogger()

const (
	SessionKey = "access_token"
)

type Controller struct {
	beego.Controller
	BasePath string
	OAuth    *auth.OAuth2
}

func MakeController(basePath string, oauth *auth.OAuth2) *Controller {
	return &Controller{
		BasePath: basePath,
		OAuth:    oauth,
	}
}

func (ac *Controller) Callback() {
	authCode := ac.Ctx.Input.Query("code")
	token, _, err := ac.OAuth.GetTokenClient(context.Background(), authCode)
	if err != nil {
		ac.CustomAbort(500, fmt.Sprintf("%+v", err))
		log.Error(fmt.Sprintf("%+v", err))
		return
	}

	ac.Ctx.Output.Session(SessionKey, token)
	path := ac.getRedirectPath()
	ac.Redirect(path, 302)
}

func (ac *Controller) getRedirectPath() string {
	requestPath := ac.Ctx.Input.Session("request_path")
	if requestPath == nil {
		return fmt.Sprintf("%s/admin/registry/overview", ac.BasePath)
	}
	ac.Ctx.Output.Session("request_path", nil)
	return requestPath.(string)
}
