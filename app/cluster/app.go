package cluster

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type JenkinsService interface {
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
}

type App struct {
	router              router.Interface
	codebaseService     codebase.ServiceInterface
	jenkinsService      jenkins.ServiceInterface
	k8sService          k8s.ServiceInterface
	gerritService       gerrit.ServiceInterface
	edpComponentService edpComponent.ServiceInterface
	codebaseName        string
	repo                string
	backupSecretName    string
}

func Make(router router.Interface, services *config.Services, cnf *config.Settings) (*App, error) {
	if !strings.Contains(cnf.Host, "//") {
		return nil, errors.New("wrong git repo")
	}

	app := App{
		router:              router,
		codebaseService:     services.Codebase,
		codebaseName:        cnf.ClusterCodebaseName,
		repo:                fmt.Sprintf("%s/%s", cnf.RegistryRepoHost, cnf.ClusterRepo),
		jenkinsService:      services.Jenkins,
		edpComponentService: services.EDPComponent,
		backupSecretName:    cnf.BackupSecretName,
		k8sService:          services.K8S,
		gerritService:       services.Gerrit,
	}

	app.createRoutes()

	return &app, nil
}