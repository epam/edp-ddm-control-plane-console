package dashboard

import (
	"context"
	"ddm-admin-console/router"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) auth(ctx *gin.Context) (response *router.Response, retErr error) {
	authCode := ctx.Query("code")
	token, _, err := a.oauth.GetTokenClient(context.Background(), authCode)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get token client")
	}

	session := sessions.Default(ctx)
	session.Set(router.AuthTokenSessionKey, token)
	if err := session.Save(); err != nil {
		return nil, errors.Wrap(err, "unable to save session")
	}

	userCtx := a.router.ContextWithUserAccessToken(ctx)

	user, err := a.openShiftService.GetMe(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get open shift user")
	}

	session = sessions.Default(ctx)
	session.Set(router.UserNameSessionKey, user.FullName)

	k8sSevice, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s service for user")
	}

	canGetClusterCodebase, err := k8sSevice.CanI("codebase", "get", a.clusterCodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}
	session.Set(router.CanViewClusterManagementSessionKey, canGetClusterCodebase)

	canListCodebases, err := k8sSevice.CanI("codebase", "list", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to codebases list")
	}
	session.Set(router.CanViewRegistriesSessionKey, canListCodebases)

	if err := session.Save(); err != nil {
		return nil, errors.Wrap(err, "unable to save session")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}

func (a *App) logout(ctx *gin.Context) (*router.Response, error) {
	session := sessions.Default(ctx)
	session.Clear()

	if err := session.Save(); err != nil {
		return nil, errors.Wrap(err, "unable to save session")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/"), nil
}
