package router

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var ConsoleVersion = "0"

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type Router struct {
	engine      *gin.Engine
	logger      Logger
	buildTime   time.Time
	appName     string
	logoMain    template.HTML
	logoFavicon string
}

type HTMLResponse struct {
	StatusResponse
	viewTemplate string
	params       gin.H
}

type StatusResponse struct {
	StatusCode int
}

func (s *StatusResponse) Code() int {
	return s.StatusCode
}

type RedirectResponse struct {
	StatusResponse
	path string
}

type JSONResponse struct {
	StatusResponse
	data interface{}
}

type Response interface {
	Code() int
}

func MakeJSONResponse(code int, data interface{}) Response {
	return &JSONResponse{
		StatusResponse: StatusResponse{StatusCode: code},
		data:           data,
	}
}

func MakeHTMLResponse(code int, viewTemplate string, params gin.H) Response {
	params["consoleVersion"] = ConsoleVersion

	return &HTMLResponse{
		StatusResponse: StatusResponse{
			StatusCode: code,
		},
		viewTemplate: viewTemplate,
		params:       params,
	}
}

func MakeStatusResponse(code int) Response {
	return &StatusResponse{StatusCode: code}
}

func MakeRedirectResponse(code int, path string) Response {
	return &RedirectResponse{
		StatusResponse: StatusResponse{StatusCode: code},
		path:           path,
	}
}

func (r *Router) makeViewResponder(handler func(ctx *gin.Context) (Response, error)) func(ctx *gin.Context) {
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

		switch rspType := rsp.(type) {
		case *StatusResponse:
			ctx.Status(rsp.Code())
		case *RedirectResponse:
			ctx.Redirect(rspType.Code(), rspType.path)
		case *HTMLResponse:
			rspType.params["appName"] = r.appName
			rspType.params = r.parseValidationErrors(rspType.params)
			rspType.params = r.includeSessionVars(ctx, rspType.params)
			rspType.params = r.includeLogos(rspType.params)
			ctx.HTML(rspType.Code(), rspType.viewTemplate, rspType.params)
		case *JSONResponse:
			ctx.JSON(rspType.Code(), rspType.data)
		}
	}
}

func (r *Router) includeSessionVars(ctx *gin.Context, params gin.H) gin.H {
	params["username"] = ctx.GetString(UserNameSessionKey)
	params["buildDate"] = r.buildTime.Unix()
	params["canViewRegistries"] = ctx.GetBool(CanViewRegistriesSessionKey) || ctx.GetBool(CanCreateRegistriesSessionKey)
	params["canViewClusterManagement"] = ctx.GetBool(CanViewClusterManagementSessionKey)
	params["canViewClusterManagement"] = ctx.GetBool(CanViewClusterManagementSessionKey)

	return params
}

func (r *Router) includeLogos(params gin.H) gin.H {
	params["logoMainSVG"] = r.logoMain
	params["logoFavicon"] = r.logoFavicon

	return params
}

func (r *Router) GET(relativePath string, handler func(ctx *gin.Context) (Response, error)) {
	r.engine.GET(relativePath, r.makeViewResponder(handler))
}

func (r *Router) POST(relativePath string, handler func(ctx *gin.Context) (Response, error)) {
	r.engine.POST(relativePath, r.makeViewResponder(handler))
}

func Make(engine *gin.Engine, logger Logger, buildTime time.Time, appName string, logoMain template.HTML, logoFavicon string) *Router {
	return &Router{
		engine:      engine,
		logger:      logger,
		buildTime:   buildTime,
		appName:     appName,
		logoMain:    logoMain,
		logoFavicon: logoFavicon,
	}
}
