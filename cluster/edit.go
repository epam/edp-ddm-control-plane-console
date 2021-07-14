package cluster

import (
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/gin-gonic/gin"
)

const (
	StorageLocation    = "backup-s3-like-storage-location"
	StorageType        = "backup-s3-like-storage-type"
	StorageCredentials = "backup-s3-like-storage-credentials"
)

type BackupConfig struct {
	StorageLocation    string `form:"storage-location" binding:"required"`
	StorageType        string `form:"storage-type" binding:"required"`
	StorageCredentials string `form:"storage-credentials" binding:"required"`
}

func (a *App) editGet(ctx *gin.Context) (*router.Response, error) {
	k8sService, err := a.k8sService.ServiceForContext(a.router.ContextWithUserAccessToken(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.codebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}
	if !canUpdateCluster {
		return nil, errors.Wrap(err, "access denied")
	}

	var backupConfig BackupConfig
	secret, err := k8sService.GetSecret(a.backupSecretName)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "unable to get backup config secret")
	}
	if err == nil {
		backupConfig = BackupConfig{
			StorageLocation:    string(secret.Data[StorageLocation]),
			StorageType:        string(secret.Data[StorageType]),
			StorageCredentials: string(secret.Data[StorageCredentials]),
		}
	}

	return router.MakeResponse(200, "cluster/edit.html", gin.H{
		"backupConf": backupConfig,
		"page":       "cluster",
	}), nil
}

func (a *App) editPost(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.codebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}
	if !canUpdateCluster {
		return nil, errors.Wrap(err, "access denied")
	}

	jenkinsService, err := a.jenkinsService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init jenkins client")
	}

	var backupConfig BackupConfig
	if err := ctx.ShouldBind(&backupConfig); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "cluster/edit.html",
			gin.H{"page": "cluster", "errorsMap": validationErrors, "backupConf": backupConfig}), nil
	}

	if err := k8sService.RecreateSecret(a.backupSecretName, map[string][]byte{
		StorageLocation:    []byte(backupConfig.StorageLocation),
		StorageType:        []byte(backupConfig.StorageType),
		StorageCredentials: []byte(backupConfig.StorageCredentials),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to recreate backup secret")
	}

	if err := jenkinsService.CreateJobBuildRun(fmt.Sprintf("cluster-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", a.codebaseName, a.codebaseName), nil); err != nil {
		return nil, errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}
