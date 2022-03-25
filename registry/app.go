package registry

import (
	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"

	"github.com/pkg/errors"
)

type App struct {
	router                     router.Interface
	codebaseService            codebase.ServiceInterface
	gerritService              gerrit.ServiceInterface
	edpComponentService        edpComponent.ServiceInterface
	k8sService                 k8s.ServiceInterface
	gerritCreatorSecretName    string
	gerritRegistryPrefix       string
	gerritRegistryHost         string
	jenkinsService             jenkins.ServiceInterface
	timezone                   string
	hardwareINITemplatePath    string
	keycloakService            keycloak.ServiceInterface
	usersRealm, usersNamespace string
}

func Make(router router.Interface, services *config.Services, cnf *config.Settings) (*App, error) {
	app := &App{
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
		keycloakService:         services.Keycloak,
		usersRealm:              cnf.UsersRealm,
		usersNamespace:          cnf.UsersNamespace,
	}

	app.createRoutes()
	if err := app.registerCustomValidators(); err != nil {
		return nil, errors.Wrap(err, "unable to register validators")
	}

	return app, nil
}
