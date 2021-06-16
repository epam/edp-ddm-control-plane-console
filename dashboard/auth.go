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

	//TODO: get user from openshift and put it to session
	session := sessions.Default(ctx)
	session.Set(a.authTokenSessionKey, token)
	if err := session.Save(); err != nil {
		return nil, errors.Wrap(err, "unable to save session")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}
