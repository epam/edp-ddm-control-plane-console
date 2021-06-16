package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type Router struct {
	engine              *gin.Engine
	logger              Logger
	authTokenSessionKey string
}

type Response struct {
	code         int
	viewTemplate string
	params       gin.H
	isRedirect   bool
}

func MakeResponse(code int, viewTemplate string, params gin.H) *Response {
	return &Response{
		code:         code,
		viewTemplate: viewTemplate,
		params:       params,
	}
}

func MakeRedirectResponse(code int, path string) *Response {
	return &Response{
		code:         code,
		viewTemplate: path,
		isRedirect:   true,
	}
}

func (r *Router) makeViewResponder(handler func(ctx *gin.Context) (*Response, error)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		rsp, err := handler(ctx)
		if err != nil {
			r.logger.Error(fmt.Sprintf("%+v", err))
			ctx.String(500, "%+v", err)
			return
		}

		if rsp == nil {
			r.logger.Error("empty view response")
			ctx.String(500, "empty view response")
			return
		}

		if rsp.isRedirect {
			ctx.Redirect(rsp.code, rsp.viewTemplate)
			return
		}

		rsp.params = r.parseValidationErrors(rsp.params)

		ctx.HTML(rsp.code, rsp.viewTemplate, rsp.params)
	}
}

func (r *Router) GET(relativePath string, handler func(ctx *gin.Context) (*Response, error)) {
	r.engine.GET(relativePath, r.makeViewResponder(handler))
}

func (r *Router) POST(relativePath string, handler func(ctx *gin.Context) (*Response, error)) {
	r.engine.POST(relativePath, r.makeViewResponder(handler))
}

func Make(engine *gin.Engine, logger Logger, authTokenSessionKey string) *Router {
	return &Router{
		engine:              engine,
		logger:              logger,
		authTokenSessionKey: authTokenSessionKey,
	}
}
