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
	ctx "ddm-admin-console/console"
	"ddm-admin-console/service/logger"
	"fmt"
	"github.com/astaxie/beego"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

type Controller struct {
	beego.Controller
}

func (ac *Controller) Callback() {
	authConfig := ctx.GetAuthConfig()
	log.Info("Start callback flow...")
	queryState := ac.Ctx.Input.Query("state")
	log.Info("State has been retrieved from query param", zap.String("queryState", queryState))
	sessionState := ac.Ctx.Input.Session(authConfig.StateAuthKey)
	log.Info("State has been retrieved from the session", zap.Any("sessionState", sessionState))
	if queryState != sessionState {
		log.Info("State does not match")
		ac.Abort("400")
		return
	}

	authCode := ac.Ctx.Input.Query("code")
	log.Info("Authorization code has been retrieved from query param")
	token, err := authConfig.Oauth2Config.Exchange(context.Background(), authCode)

	if err != nil {
		log.Info("Failed to exchange token with code", zap.String("code", authCode))
		ac.Abort("500")
		return
	}
	log.Info("Authorization code has been successfully exchanged with token")

	ts := authConfig.Oauth2Config.TokenSource(context.Background(), token)

	ac.Ctx.Output.Session("token_source", ts)
	log.Info("Token source has been saved to the session")
	path := ac.getRedirectPath()
	ac.Redirect(path, 302)
}

func (ac *Controller) getRedirectPath() string {
	requestPath := ac.Ctx.Input.Session("request_path")
	if requestPath == nil {
		return fmt.Sprintf("%s/admin/edp/overview", ctx.BasePath)
	}
	ac.Ctx.Output.Session("request_path", nil)
	return requestPath.(string)
}
