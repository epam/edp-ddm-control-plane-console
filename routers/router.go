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

package routers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/controllers"
	"ddm-admin-console/controllers/auth"
	"ddm-admin-console/filters"
	"ddm-admin-console/k8s"
	"ddm-admin-console/repository"
	edpComponentRepo "ddm-admin-console/repository/edp-component"
	"ddm-admin-console/service"
	cbs "ddm-admin-console/service/codebasebranch"
	edpComponentService "ddm-admin-console/service/edp-component"
	"ddm-admin-console/service/logger"
	"ddm-admin-console/util"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

const (
	integrationStrategies = "integrationStrategies"
	buildTools            = "buildTools"
	versioningTypes       = "versioningTypes"
	testReportTools       = "testReportTools"
	deploymentScript      = "deploymentScript"
	ciTools               = "ciTools"
)

func init() {
	log.Info("Start application...",
		zap.String("mode", beego.AppConfig.String("runmode")))
	authEnabled, err := beego.AppConfig.Bool("keycloakAuthEnabled")
	if err != nil {
		log.Error("Cannot read property keycloakAuthEnabled. Set default: true", zap.Error(err))
		authEnabled = true
	}

	if authEnabled {
		console.InitAuth()

		beego.Router(fmt.Sprintf("%s/auth/callback", console.BasePath), &auth.Controller{}, "get:Callback")
		beego.InsertFilter(fmt.Sprintf("%s/admin/*", console.BasePath), beego.BeforeRouter, filters.AuthFilter)
		beego.InsertFilter(fmt.Sprintf("%s/api/v1/edp/*", console.BasePath), beego.BeforeRouter, filters.AuthRestFilter)
		beego.InsertFilter(fmt.Sprintf("%s/admin/*", console.BasePath), beego.BeforeRouter, filters.RoleAccessControlFilter)
		beego.InsertFilter(fmt.Sprintf("%s/api/v1/edp/*", console.BasePath), beego.BeforeRouter, filters.RoleAccessControlRestFilter)
	} else {
		beego.InsertFilter(fmt.Sprintf("%s/*", console.BasePath), beego.BeforeRouter, filters.StubAuthFilter)
	}

	clients := k8s.CreateOpenShiftClients()
	codebaseRepository := repository.CodebaseRepository{}
	branchRepository := repository.CodebaseBranchRepository{}
	ecr := edpComponentRepo.EDPComponent{}

	ecs := edpComponentService.Service{IEDPComponent: ecr}
	edpService := service.EDPTenantService{Clients: clients}
	branchService := cbs.Service{
		Clients:                  clients,
		IReleaseBranchRepository: branchRepository,
		ICodebaseRepository:      codebaseRepository,
		CodebaseBranchValidation: map[string]func(string, string) ([]string, error){},
	}
	codebaseService := service.CodebaseService{
		Clients:                 clients,
		ICodebaseRepository:     codebaseRepository,
		BranchService:           branchService,
		GerritCreatorSecretName: beego.AppConfig.String("gerritCreatorSecretName"),
	}

	ec := controllers.EDPTenantController{
		EDPTenantService: edpService,
		EDPComponent:     ecs,
	}

	beego.ErrorController(&controllers.ErrorController{})
	beego.Router(fmt.Sprintf("%s/", console.BasePath), &controllers.MainController{EDPTenantService: edpService}, "get:Index")
	beego.SetStaticPath(fmt.Sprintf("%s/static", console.BasePath), "static")

	integrationStrategies := util.GetValuesFromConfig(integrationStrategies)
	if integrationStrategies == nil {
		log.Fatal("integrationStrategies config variable is empty.")
	}

	buildTools := util.GetValuesFromConfig(buildTools)
	if buildTools == nil {
		log.Fatal("buildTools config variable is empty.")
	}

	vt := util.GetValuesFromConfig(versioningTypes)
	if vt == nil {
		log.Fatal("versioningTypes config variable is empty.")
	}

	testReportTools := util.GetValuesFromConfig(testReportTools)
	if testReportTools == nil {
		log.Fatal("testReportTools config variable is empty.")
	}

	ds := util.GetValuesFromConfig(deploymentScript)
	if ds == nil {
		log.Fatal("deploymentScript config variable is empty.")
	}

	ciTools := util.GetValuesFromConfig(ciTools)
	if ciTools == nil {
		log.Fatal("ciTools config variable is empty.")
	}

	is := make([]string, len(integrationStrategies))
	copy(is, integrationStrategies)

	autis := make([]string, len(integrationStrategies))
	copy(autis, integrationStrategies)

	k8sEDPComponentService := edpComponentService.MakeServiceK8S(clients.EDPRestClientV1)

	adminEdpNamespace := beego.NewNamespace(fmt.Sprintf("%s/admin", console.BasePath),
		beego.NSRouter("/overview", &ec, "get:GetEDPComponents"),
		beego.NSRouter("/dashboard", controllers.MakeDashboardController()),

		beego.NSRouter("/registry/overview", controllers.MakeListRegistry(&codebaseService)),
		beego.NSRouter("/registry/create", controllers.MakeCreateRegistry(&codebaseService)),
		beego.NSRouter("/registry/edit/:name", controllers.MakeEditRegistry(&codebaseService)),
		beego.NSRouter("/registry/view/:name", controllers.MakeViewRegistry(&codebaseService, k8sEDPComponentService)),
		beego.NSRouter("/cluster/management", controllers.MakeClusterManagement(&codebaseService,
			k8sEDPComponentService,
			beego.AppConfig.DefaultString("clusterManagementCodebaseName", "cluster-management"),
			beego.AppConfig.String("clusterManagementRepo"))),
	)
	beego.AddNamespace(adminEdpNamespace)

	if err := i18n.SetMessage("uk", "conf/locale_uk-UA.ini"); err != nil {
		log.Fatal(err.Error())
	}
	if err := beego.AddFuncMap("i18n", i18n.Tr); err != nil {
		log.Fatal(err.Error())
	}
}
