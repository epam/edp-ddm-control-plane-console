package router

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	UserTokenKey = "access-token"
)

func (r *Router) ContextWithUserAccessToken(ctx *gin.Context) context.Context {
	session := sessions.Default(ctx)
	token := session.Get(r.authTokenSessionKey)
	if token == nil {
		return context.Background()
	}

	tokenData, ok := token.(*oauth2.Token)
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), UserTokenKey, tokenData.AccessToken)
}
