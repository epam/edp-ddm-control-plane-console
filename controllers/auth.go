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

	"github.com/astaxie/beego"
	bgCtx "github.com/astaxie/beego/context"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AuthTokenSessionKey = "access_token"
)

type AuthController struct {
	beego.Controller
	BasePath, Namespace string
	OAuth               *auth.OAuth2
	K8SClients          *k8s.ClientSet
}

func MakeAuthController(basePath, namespace string, oauth *auth.OAuth2, k8sClients *k8s.ClientSet) *AuthController {
	return &AuthController{
		BasePath:   basePath,
		Namespace:  namespace,
		OAuth:      oauth,
		K8SClients: k8sClients,
	}
}

func (ac *AuthController) Callback() {
	err := ac.parseCallback()
	if err != nil {
		ac.CustomAbort(500, fmt.Sprintf("%+v", err))
		log.Error(fmt.Sprintf("%+v", err))
		return
	}

	path := ac.getRedirectPath()
	ac.Redirect(path, 302)
}

func (ac *AuthController) parseCallback() error {
	authCode := ac.Ctx.Input.Query("code")
	token, _, err := ac.OAuth.GetTokenClient(context.Background(), authCode)
	if err != nil {
		return errors.Wrap(err, "unable to get token client")
	}

	ac.Ctx.Output.Session(AuthTokenSessionKey, token)

	userCl, err := ac.K8SClients.GetUserClient(contextWithUserAccessToken(ac.Ctx))
	if err != nil {
		return errors.Wrap(err, "unable to init user client")
	}

	me, err := userCl.Users().Get("~", v12.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to read current user info")
	}

	rbacClient, err := ac.K8SClients.GetRbacClient(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to ini rbac client")
	}

	bindings, err := rbacClient.RoleBindings(ac.Namespace).List(v12.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to get role bindings")
	}

	var userRoles []string
	for _, b := range bindings.Items {
		for _, s := range b.Subjects {
			if s.Kind == "User" && s.Name == me.Name && b.RoleRef.Kind == "Role" {
				userRoles = append(userRoles, b.RoleRef.Name)
			}
		}
	}
	ac.Ctx.Output.Session(sessionGroupsKey, userRoles)

	return nil
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
