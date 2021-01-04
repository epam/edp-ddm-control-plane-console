# Admin Console

## Overview
Admin Console is a central management tool in the EDP ecosystem that provides 
the ability to deploy infrastructure, 
project resources and new technologies simple.

## Dependencies
- go v1.14
- github.com/astaxie/beego v1.12.0
- github.com/docker/docker v1.13.1
- github.com/epmd-edp/codebase-operator/v2 v2.3.0
- github.com/epmd-edp/cd-pipeline-operator/v2 v2.3.0
- github.com/golang-migrate/migrate v3.5.4
- github.com/lib/pq v1.0.0
- github.com/openshift/api v3.9.0
- github.com/openshift/client-go v3.9.0
- github.com/pkg/errors v0.8.1
- go.uber.org/zap v1.14.1
- k8s.io/api
- k8s.io/apimachinery
- k8s.io/client-go

## Setup for local development
1. This project uses **go modules** so to install all dependencies you need to run
`go mod download`
2. Configure environment, database connection and other parameters in conf/app.conf
3. Compile and run

## Errors & Logging
The project uses pkg/errors lib to handle errors and stack traces. Every error must be wrapped as follows:
```
if err != nil {
  return nil, errors.Wrap(err, "unable to get registry list")
}
```
Wraping an error added context to the underlying error and recorded the file and 
line that the error occurred. This file and line information could be retrieved via a 
helper function, Fprint, to give a trace of the execution path leading away from the error.

**zap** - Blazing fast, structured, leveled logging in Go.
To use logging you need to get logger instance from `service/logger` package:
```
var log = logger.GetLogger()
```
There is different logging level functions provided by logger interface:
- **Info** - Info logs a message at InfoLevel;
- **Warn** - Warn logs a message at WarnLevel;
- **Error** - Error logs a message at ErrorLevel;
- **Debug** - Debug logs a message at DebugLevel;
- **DPanic** - DPanic logs a message at DPanicLevel. If the logger is in development mode, it then panics 
(DPanic means "development panic"). This is useful for catching errors that are recoverable, but shouldn't ever happen;
- **Panic** - Panic logs a message at PanicLevel. The logger then panics, even if logging at PanicLevel is disabled;
- **Fatal** - Fatal logs a message at FatalLevel; The logger then calls os.Exit(1), 
even if logging at FatalLevel is disabled.

## HTTP Handling & Routing
Project uses **beego** framework to handle http requests, route traffic, 
perform DB connections and render templates. More details you can find here [https://beego.me/](https://beego.me/)

- [Routing](https://beego.me/docs/mvc/controller/router.md)
- [ORM](https://beego.me/docs/mvc/model/overview.md)
- [Template Parsing](https://beego.me/docs/mvc/view/view.md)
- [i18n](https://beego.me/docs/module/i18n.md)

## DB Migrations
Database migrations is handled by [github.com/golang-migrate/migrate](https://github.com/golang-migrate/migrate) library.

- Migrate reads migrations from sources and applies them in correct order to a database.
- Drivers are "dumb", migrate glues everything together and makes sure the logic is bulletproof. (Keeps the drivers lightweight, too.)
- Database drivers don't assume things or try to correct user input. When in doubt, fail.

For local development you should follow instruction, how to setup migrate cli:

https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

Then you can use this command to verify your migration scripts:

`migrate -path=go/src/ddm-admin-console/db/migrations -database="postgres://user:password@localhost:5432/edp?sslmode=disable&search_path=main" goto 71`

Migrate is integrated into project structure so migrations will run automatically on project startup.
If you want to write a new migration please see `db/migrations` folder.

## Project structure
- conf - folder with configuration files
- controllers - beego controllers
- deploy-templates - helm charts
- db - database migrations
- filters - beego http request filters - auth, rbac, etc..
- k8s - core k8s/openshift client and interface
- models - model structures with tag annotations
- repository - structures and interfaces for querying database
- routers - global init package with basic router and controllers initialization
- static - static files css, js images
- templatefunction - function helpers for html rendering
- test - additional test utils and mocks
- views - html templates