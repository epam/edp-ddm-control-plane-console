package dashboard

import (
	"context"
	"ddm-admin-console/auth"
	"ddm-admin-console/router"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/openshift"

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
	GetAll() ([]edpComponent.EDPComponent, error)
}

type App struct {
	router              Router
	logger              Logger
	edpComponentService EDPComponentService
	oauth               *auth.OAuth2 //TODO: interface
	k8sService          k8s.ServiceInterface
	openShiftService    openshift.ServiceInterface

	clusterCodebaseName string
}

func Make(router Router, edpComponentService EDPComponentService, oauth *auth.OAuth2,
	k8sService k8s.ServiceInterface, openShiftService openshift.ServiceInterface, clusterCodebaseName string) (*App, error) {
	app := App{
		router:              router,
		edpComponentService: edpComponentService,
		oauth:               oauth,
		k8sService:          k8sService,
		openShiftService:    openShiftService,
		clusterCodebaseName: clusterCodebaseName,
	}

	app.createRoutes()

	return &app, nil
}
