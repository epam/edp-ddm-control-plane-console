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
	oauth "ddm-admin-console/auth"
	"ddm-admin-console/controllers"
	"ddm-admin-console/k8s"
	"ddm-admin-console/repository"
	edpComponentRepo "ddm-admin-console/repository/edp-component"
	"ddm-admin-console/service"
	edpComponentService "ddm-admin-console/service/edp-component"
	"ddm-admin-console/service/logger"
	"ddm-admin-console/util"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

const (
	buildTools       = "buildTools"
	versioningTypes  = "versioningTypes"
	testReportTools  = "testReportTools"
	deploymentScript = "deploymentScript"
	ciTools          = "ciTools"
)

func init() {
	var (
		log          = logger.GetLogger()
		basePath     = beego.AppConfig.String("basePath")
		tenant       = beego.AppConfig.String("edpName")
		namespace    = beego.AppConfig.String("namespace")
		host         = beego.AppConfig.String("host")
		creatorGroup = beego.AppConfig.String("creatorGroup")
	)

	log.Info("Start application...",
		zap.String("mode", beego.AppConfig.String("runmode")))
	authEnabled, err := beego.AppConfig.Bool("authEnabled")
	if err != nil {
		log.Error("Cannot read property authEnabled. Set default: true", zap.Error(err))
		authEnabled = true
	}

	k8sClients, err := k8s.MakeK8SClients()
	if err != nil {
		panic(err)
	}

	if authEnabled {
		transport, err := rest.TransportFor(k8sClients.GetConfig())
		if err != nil {
			panic(err)
		}

		oa, err := oauth.InitOauth2(
			beego.AppConfig.String("clientId"),
			beego.AppConfig.String("clientSecret"),
			k8sClients.GetConfig().Host,
			host+basePath+"/auth/callback",
			&http.Client{Transport: transport})
		if err != nil {
			panic(err)
		}

		beego.Router(fmt.Sprintf("%s/auth/callback", basePath),
			controllers.MakeAuthController(basePath, namespace, oa, k8sClients), "get:Callback")
		beego.InsertFilter(fmt.Sprintf("%s/admin/*", basePath), beego.BeforeRouter,
			oauth.MakeBeegoFilter(oa, controllers.AuthTokenSessionKey))
	}

	codebaseRepository := repository.CodebaseRepository{}
	branchRepository := repository.CodebaseBranchRepository{}
	ecr := edpComponentRepo.EDPComponent{}

	ecs := edpComponentService.MakeService(ecr, namespace)
	edpService := service.EDPTenantService{}
	branchService := service.MakeCodebaseBranchService(k8sClients, branchRepository, codebaseRepository, namespace)

	codebaseService := service.MakeCodebaseService(k8sClients,
		codebaseRepository, branchService, beego.AppConfig.String("gerritCreatorSecretName"), namespace)

	ec := controllers.MakeEDPTenantController(tenant, basePath, &edpService, ecs)

	beego.ErrorController(controllers.MakeErrorController(basePath))
	beego.Router(fmt.Sprintf("%s/", basePath), controllers.MakeMainController(basePath, &edpService), "get:Index")
	beego.SetStaticPath(fmt.Sprintf("%s/static", basePath), "static")

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

	k8sEDPComponentService := edpComponentService.MakeServiceK8S(k8sClients.EDPRestClientV1)

	projectsSvc := service.MakeProjects(k8sClients)
	jenkinsSvc := service.MakeJenkins(k8sClients, namespace)

	adminEdpNamespace := beego.NewNamespace(fmt.Sprintf("%s/admin", basePath),
		beego.NSRouter("/overview", ec, "get:GetEDPComponents"),
		beego.NSRouter("/dashboard", controllers.MakeDashboardController()),

		beego.NSRouter("/registry/overview", controllers.MakeListRegistry(basePath, creatorGroup,
			codebaseService, projectsSvc)),
		beego.NSRouter("/registry/create", controllers.MakeCreateRegistry(basePath, creatorGroup, codebaseService)),
		beego.NSRouter("/registry/edit/:name", controllers.MakeEditRegistry(codebaseService, projectsSvc,
			jenkinsSvc)),
		beego.NSRouter("/registry/view/:name", controllers.MakeViewRegistry(codebaseService,
			k8sEDPComponentService, projectsSvc, basePath, namespace)),
		beego.NSRouter("/cluster/management", controllers.MakeClusterManagement(jenkinsSvc, codebaseService,
			k8sEDPComponentService,
			beego.AppConfig.DefaultString("clusterManagementCodebaseName", "cluster-management"),
			beego.AppConfig.String("clusterManagementRepo"), basePath, namespace)),
	)
	beego.AddNamespace(adminEdpNamespace)

	if err := i18n.SetMessage("uk", "conf/locale_uk-UA.ini"); err != nil {
		log.Fatal(err.Error())
	}
	if err := beego.AddFuncMap("i18n", i18n.Tr); err != nil {
		log.Fatal(err.Error())
	}
}
