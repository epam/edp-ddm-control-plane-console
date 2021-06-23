package router

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	AuthTokenSessionKey                = "access-token"
	AuthTokenValidSessionKey           = "access-token-valid"
	UserNameSessionKey                 = "user-full-name"
	CanViewClusterManagementSessionKey = "can-view-cluster-management"
	CanViewRegistriesSessionKey        = "can-view-registries"
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

func UserDataMiddleware(ctx *gin.Context) {
	tokenIsValid := ctx.GetBool(AuthTokenValidSessionKey)
	if !tokenIsValid {
		ctx.Next()
		return
	}

	session := sessions.Default(ctx)

	sessVal := session.Get(UserNameSessionKey)
	if sessVal != nil {
		val, ok := sessVal.(string)
		if ok {
			ctx.Set(UserNameSessionKey, val)
		}
	}

	sessVal = session.Get(CanViewClusterManagementSessionKey)
	if sessVal != nil {
		val, ok := sessVal.(bool)
		if ok {
			ctx.Set(CanViewClusterManagementSessionKey, val)
		}
	}

	sessVal = session.Get(CanViewRegistriesSessionKey)
	if sessVal != nil {
		val, ok := sessVal.(bool)
		if ok {
			ctx.Set(CanViewRegistriesSessionKey, val)
		}
	}

	ctx.Next()
}
