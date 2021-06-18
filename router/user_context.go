package router

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	AuthTokenSessionKey = "access-token"
	UserNameSessionKey  = "user-full-name"
)

func (r *Router) ContextWithUserAccessToken(ctx *gin.Context) context.Context {
	session := sessions.Default(ctx)
	token := session.Get(AuthTokenSessionKey)
	if token == nil {
		return context.Background()
	}

	tokenData, ok := token.(*oauth2.Token)
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), AuthTokenSessionKey, tokenData.AccessToken)
}

func UserNameMiddleware(ctx *gin.Context) {
	session := sessions.Default(ctx)

	userName := session.Get(UserNameSessionKey)
	if userName == nil {
		return
	}

	val, ok := userName.(string)
	if !ok {
		return
	}

	ctx.Set(UserNameSessionKey, val)
}
