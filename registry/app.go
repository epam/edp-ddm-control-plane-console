package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/k8s"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type Router interface {
	GET(relativePath string, handler func(ctx *gin.Context) (*router.Response, error))
	POST(relativePath string, handler func(ctx *gin.Context) (*router.Response, error))
	ContextWithUserAccessToken(ctx *gin.Context) context.Context
}

type EDPComponentService interface {
	Get(name string) (*edpComponent.EDPComponent, error)
	GetAllNamespace(ns string) ([]edpComponent.EDPComponent, error)
}

type JenkinsService interface {
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
}

type App struct {
	router                  Router
	logger                  Logger
	codebaseService         codebase.ServiceInterface
	edpComponentService     EDPComponentService
	k8sService              k8s.ServiceInterface
	registryGitRepo         string
	gerritCreatorSecretName string
	jenkinsService          JenkinsService
}

func Make(router Router, logger Logger, codebaseService codebase.ServiceInterface, edpComponentService EDPComponentService,
	k8sService k8s.ServiceInterface, jenkinsService JenkinsService, registryGitRepo, gerritCreatorSecretName string) (*App, error) {
	app := &App{
		logger:                  logger,
		router:                  router,
		codebaseService:         codebaseService,
		edpComponentService:     edpComponentService,
		k8sService:              k8sService,
		registryGitRepo:         registryGitRepo,
		gerritCreatorSecretName: gerritCreatorSecretName,
		jenkinsService:          jenkinsService,
	}

	app.createRoutes()
	if err := app.registerCustomValidators(); err != nil {
		return nil, errors.Wrap(err, "unable to register validators")
	}

	return app, nil
}
