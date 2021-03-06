package cluster

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/vault"
	"fmt"

	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type JenkinsService interface {
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
}

type Services struct {
	Codebase     codebase.ServiceInterface
	Jenkins      jenkins.ServiceInterface
	K8S          k8s.ServiceInterface
	Gerrit       gerrit.ServiceInterface
	EDPComponent edpComponent.ServiceInterface
	Vault        vault.ServiceInterface
}

type Config struct {
	CodebaseName                   string
	BackupSecretName               string
	RegistryRepoHost               string
	ClusterRepo                    string
	VaultClusterAdminsPathTemplate string
	VaultClusterAdminsPasswordKey  string
	VaultKVEngineName              string
	HardwareINITemplatePath        string
}

type App struct {
	Services
	Config
	router router.Interface
	repo   string
}

func Make(router router.Interface, services Services, cnf Config) *App {
	app := App{
		Services: services,
		Config:   cnf,
		router:   router,
		repo:     fmt.Sprintf("%s/%s", cnf.RegistryRepoHost, cnf.ClusterRepo),
	}

	app.createRoutes()

	return &app
}
