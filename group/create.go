package group

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
)

type createView struct {
	app *App
}

func (c createView) Get(ctx *gin.Context) (*router.Response, error) {
	userCtx := c.app.router.ContextWithUserAccessToken(ctx)
	k8sService, err := c.app.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := c.app.codebaseService.CheckIsAllowedToCreate(k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check create access")
	}

	if !allowed {
		return nil, errors.New("access denied")
	}

	return router.MakeResponse(200, "group/create.html", gin.H{
		"page": "group",
	}), nil
}

func (c createView) Post(ctx *gin.Context) (*router.Response, error) {
	var g group
	if err := ctx.ShouldBind(&g); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse group form")
		}

		return router.MakeResponse(200, "group/create.html",
			gin.H{"page": "group", "errorsMap": validationErrors, "model": g}), nil
	}

	if err := c.createGroup(ctx, &g); err != nil {
		validationErrors, ok := errors.Cause(err).(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "group/create.html",
			gin.H{"page": "group", "errorsMap": validationErrors, "model": g}), nil
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/group/overview"), nil
}

func (c createView) createGroup(ctx *gin.Context, gr *group) error {
	userCtx := c.app.router.ContextWithUserAccessToken(ctx)
	k8sService, err := c.app.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := c.app.codebaseService.CheckIsAllowedToCreate(k8sService)
	if err != nil {
		return errors.Wrap(err, "access denied")
	}
	if !allowed {
		return errors.New("access denied")
	}

	cbService, err := c.app.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	_, err = cbService.Get(gr.Name)
	if err == nil {
		return validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Name", "group-exists")})
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unknown error")
	}

	cb := initGroupCodebase(gr.Name, gr.Description, c.app.groupGitRepo)

	if err := cbService.Create(cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := cbService.CreateTempSecrets(cb, k8sService, c.app.gerritCreatorSecretName); err != nil {
		return errors.Wrap(err, "unable to create codebase tmp secrets")
	}

	if err := cbService.CreateDefaultBranch(cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func initGroupCodebase(name, description, groupGitRepo string) *codebase.Codebase {
	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
	return &codebase.Codebase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       "Codebase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: codebase.CodebaseSpec{
			Description:      &description,
			Type:             codebase.GroupCodebaseType,
			BuildTool:        "gitops",
			Lang:             "other",
			DefaultBranch:    "master",
			Strategy:         "clone",
			DeploymentScript: "openshift-template",
			GitServer:        "gerrit",
			CiTool:           "Jenkins",
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				Url: groupGitRepo,
			},
			JenkinsSlave: &jenkinsSlave,
		},
		Status: codebase.CodebaseStatus{
			Available:       false,
			LastTimeUpdated: time.Now(),
			Status:          "initialized",
			Action:          "codebase_registration",
			Value:           "inactive",
		},
	}
}
