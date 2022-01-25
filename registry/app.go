package registry

import (
	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"

	"github.com/pkg/errors"

	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type App struct {
	router                  router.Interface
	logger                  Logger
	codebaseService         codebase.ServiceInterface
	gerritService           gerrit.ServiceInterface
	edpComponentService     edpComponent.ServiceInterface
	k8sService              k8s.ServiceInterface
	gerritCreatorSecretName string
	gerritRegistryPrefix    string
	gerritRegistryHost      string
	jenkinsService          jenkins.ServiceInterface
	timezone                string
	hardwareINITemplatePath string
}

func Make(router router.Interface, logger Logger, services *config.Services, cnf *config.Settings) (*App, error) {
	app := &App{
		logger:                  logger,
		router:                  router,
		codebaseService:         services.Codebase,
		edpComponentService:     services.EDPComponent,
		k8sService:              services.K8S,
		gerritCreatorSecretName: cnf.GerritCreatorSecretName,
		jenkinsService:          services.Jenkins,
		timezone:                cnf.Timezone,
		gerritService:           services.Gerrit,
		gerritRegistryPrefix:    cnf.RegistryRepoPrefix,
		gerritRegistryHost:      cnf.RegistryRepoHost,
		hardwareINITemplatePath: cnf.RegistryHardwareKeyINITemplatePath,
	}

	app.createRoutes()
	if err := app.registerCustomValidators(); err != nil {
		return nil, errors.Wrap(err, "unable to register validators")
	}

	return app, nil
}
