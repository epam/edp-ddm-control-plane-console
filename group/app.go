package group

import (
	"fmt"

	edpComponent "ddm-admin-console/service/edp_component"

	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
)

type App struct {
	router                  router.Interface
	codebaseService         codebase.ServiceInterface
	k8sService              k8s.ServiceInterface
	edpComponentService     edpComponent.ServiceInterface
	groupGitRepo            string
	gerritCreatorSecretName string
	timezone                string
}

func Make(r router.Interface, services *config.Services, cnf *config.Settings) *App {
	a := App{
		router:                  r,
		codebaseService:         services.Codebase,
		k8sService:              services.K8S,
		groupGitRepo:            fmt.Sprintf("%s/%s", cnf.RegistryRepoHost, cnf.GroupGitRepo),
		gerritCreatorSecretName: cnf.GerritCreatorSecretName,
		timezone:                cnf.Timezone,
		edpComponentService:     services.EDPComponent,
	}

	a.createRoutes()

	return &a
}

func (a *App) createRoutes() {
	a.router.AddView("/admin/group/create", createView{app: a})
	a.router.AddView("/admin/group/overview", listView{app: a})
	a.router.AddView("/admin/group/edit/:name", editView{app: a})
	a.router.GET("/admin/group/view/:name", a.detailsView)
}
