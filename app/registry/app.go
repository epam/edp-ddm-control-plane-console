package registry

import (
	"strings"

	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"

	"github.com/pkg/errors"
)

type App struct {
	router                   router.Interface
	codebaseService          codebase.ServiceInterface
	gerritService            gerrit.ServiceInterface
	edpComponentService      edpComponent.ServiceInterface
	k8sService               k8s.ServiceInterface
	gerritRegistryPrefix     string
	gerritRegistryHost       string
	jenkinsService           jenkins.ServiceInterface
	timezone                 string
	hardwareINITemplatePath  string
	EnableBranchProvisioners bool
	clusterCodebaseName      string
	codebaseLabels           map[string]string
	admins                   *Admins
}

func Make(router router.Interface, services *config.Services, cnf *config.Settings) (*App, error) {
	app := &App{
		router:                   router,
		codebaseService:          services.Codebase,
		edpComponentService:      services.EDPComponent,
		k8sService:               services.K8S,
		jenkinsService:           services.Jenkins,
		timezone:                 cnf.Timezone,
		gerritService:            services.Gerrit,
		gerritRegistryPrefix:     cnf.RegistryRepoPrefix,
		gerritRegistryHost:       cnf.RegistryRepoHost,
		hardwareINITemplatePath:  cnf.RegistryHardwareKeyINITemplatePath,
		EnableBranchProvisioners: cnf.EnableBranchProvisioners,
		clusterCodebaseName:      cnf.ClusterCodebaseName,
		admins:                   MakeAdmins(services.Keycloak, cnf.UsersRealm, cnf.UsersNamespace),
	}

	if cnf.RegistryCodebaseLabels != "" {
		labels := strings.Split(cnf.RegistryCodebaseLabels, ",")
		if len(labels) > 0 {
			app.codebaseLabels = make(map[string]string)
			for _, l := range labels {
				labelParts := strings.Split(l, "=")
				if len(labelParts) == 2 {
					app.codebaseLabels[labelParts[0]] = labelParts[1]
				}
			}
		}
	}

	app.createRoutes()
	if err := app.registerCustomValidators(); err != nil {
		return nil, errors.Wrap(err, "unable to register validators")
	}

	return app, nil
}
