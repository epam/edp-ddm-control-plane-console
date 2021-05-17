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
	"context"
	"ddm-admin-console/auth"
	"ddm-admin-console/k8s"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/astaxie/beego"
	bgCtx "github.com/astaxie/beego/context"
)

const (
	AuthTokenSessionKey = "access_token"
)

type AuthController struct {
	beego.Controller
	BasePath string
	OAuth    *auth.OAuth2
}

func MakeAuthController(basePath string, oauth *auth.OAuth2) *AuthController {
	return &AuthController{
		BasePath: basePath,
		OAuth:    oauth,
	}
}

func (ac *AuthController) Callback() {
	authCode := ac.Ctx.Input.Query("code")
	token, _, err := ac.OAuth.GetTokenClient(context.Background(), authCode)
	if err != nil {
		ac.CustomAbort(500, fmt.Sprintf("%+v", err))
		log.Error(fmt.Sprintf("%+v", err))
		return
	}

	ac.Ctx.Output.Session(AuthTokenSessionKey, token)
	path := ac.getRedirectPath()
	ac.Redirect(path, 302)
}

func (ac *AuthController) getRedirectPath() string {
	requestPath := ac.Ctx.Input.Session("request_path")
	if requestPath == nil {
		return fmt.Sprintf("%s/admin/registry/overview", ac.BasePath)
	}
	ac.Ctx.Output.Session("request_path", nil)
	return requestPath.(string)
}

//TODO: try to embed this function in all controllers
func contextWithUserAccessToken(ctx *bgCtx.Context) context.Context {
	token := ctx.Input.Session(AuthTokenSessionKey)
	if token == nil {
		return context.Background()
	}

	tokenData, ok := token.(*oauth2.Token)
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), k8s.UserTokenKey, tokenData.AccessToken)
}
