package auth

import (
	"net/http"

	"golang.org/x/oauth2"

	bgCtx "github.com/astaxie/beego/context"
)

func MakeBeegoFilter(o *OAuth2, sessionKey string) func(context *bgCtx.Context) {
	return func(context *bgCtx.Context) {
		tsRaw := context.Input.Session(sessionKey)
		if tsRaw == nil {
			beegoStartAuth(o, context)
			return
		}

		token, ok := tsRaw.(*oauth2.Token)
		if !ok {
			beegoStartAuth(o, context)
			return
		}

		if !token.Valid() {
			beegoStartAuth(o, context)
		}
	}
}

func beegoStartAuth(o *OAuth2, context *bgCtx.Context) {
	context.Redirect(http.StatusFound, o.AuthCodeURL())
}
