package cluster

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
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

type CodebaseService interface {
	Create(cb *codebase.Codebase) error
	CreateDefaultBranch(cb *codebase.Codebase) error
	Get(name string) (*codebase.Codebase, error)
	GetBranchesByCodebase(codebaseName string) ([]codebase.CodebaseBranch, error)
}

type JenkinsService interface {
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
}

type EDPComponentService interface {
	Get(name string) (*edpComponent.EDPComponent, error)
	GetAllNamespace(ns string) ([]edpComponent.EDPComponent, error)
}

type K8SService interface {
	RecreateSecret(secretName string, data map[string][]byte) error
	GetSecret(name string) (*v1.Secret, error)
}

type App struct {
	router              Router
	logger              Logger
	codebaseService     CodebaseService
	jenkinsService      JenkinsService
	k8sService          K8SService
	edpComponentService EDPComponentService

	codebaseName            string
	repo                    string
	gerritCreatorSecretName string
	backupSecretName        string
}

func Make(router Router, logger Logger, codebaseService CodebaseService, jenkinsService JenkinsService,
	edpComponentService EDPComponentService, k8sService K8SService, codebaseName, repo, gerritCreatorSecretName,
	backupSecretName string) (*App, error) {

	if !strings.Contains(repo, "//") || !strings.Contains(repo, "/") {
		return nil, errors.New("wrong git repo")
	}

	app := App{
		router:                  router,
		logger:                  logger,
		codebaseService:         codebaseService,
		codebaseName:            codebaseName,
		repo:                    repo,
		jenkinsService:          jenkinsService,
		gerritCreatorSecretName: gerritCreatorSecretName,
		edpComponentService:     edpComponentService,
		backupSecretName:        backupSecretName,
		k8sService:              k8sService,
	}

	app.createRoutes()

	return &app, nil
}
