package auth

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func MakeGinMiddleware(o *OAuth2, sessionKey, filterPath string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.Request.RequestURI, filterPath) {
			ctx.Next()
			return
		}

		session := sessions.Default(ctx)
		tsRaw := session.Get(sessionKey)
		if tsRaw == nil {
			ginStartAuth(o, ctx)
			return
		}

		token, ok := tsRaw.(*oauth2.Token)
		if !ok {
			ginStartAuth(o, ctx)
			return
		}

		if !token.Valid() {
			ginStartAuth(o, ctx)
			return
		}

		ctx.Next()
	}
}

func ginStartAuth(o *OAuth2, ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, o.AuthCodeURL())
}
