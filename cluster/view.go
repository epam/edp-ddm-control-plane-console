package cluster

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"fmt"
	"strings"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	codebaseDescription   = "Керування інфрастуктурними компонентами кластеру"
	defaultGitServer      = "gerrit"
	clusterManagementType = "clustermgmt"
	defaultBranch         = "master"
	lang                  = "other"
	buildTool             = "gitops"
	deploymentScript      = "openshift-template"
	ciTool                = "Jenkins"
	jenkinsSlave          = "gitops"
	gerritCreatorUsername = "user"
	gerritCreatorPassword = "password"
)

func (a *App) view(ctx *gin.Context) (*router.Response, error) {
	cb, err := a.codebaseService.Get(a.codebaseName)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "unable to get cluster codebase")
	}

	if err != nil && k8sErrors.IsNotFound(err) {
		cb, err = a.createClusterCodebase()
		if err != nil {
			return nil, errors.Wrap(err, "unable to create cluster codebase")
		}
	}

	branches, err := a.codebaseService.GetBranchesByCodebase(cb.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry branches")
	}
	cb.Branches = branches

	jenkinsComponent, err := a.edpComponentService.Get("jenkins")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.edpComponentService.Get("gerrit")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.edpComponentService.GetAllNamespace(cb.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaced edp components")
	}

	return router.MakeResponse(200, "cluster/view.html", gin.H{
		"branches":      branches,
		"codebase":      cb,
		"jenkinsURL":    jenkinsComponent.Spec.Url,
		"gerritURL":     gerritComponent.Spec.Url,
		"page":          "cluster",
		"edpComponents": namespacedEDPComponents,
	}), nil
}

func (a *App) createClusterCodebase() (*codebase.Codebase, error) {
	//username, _ := c.Ctx.Input.Session("username").(string)
	jobProvisioning := "default"
	startVersion := "0.0.1"
	description := codebaseDescription
	repo := fmt.Sprintf("/%s", strings.Join(strings.Split(strings.Split(a.repo, "//")[1], "/")[1:], "/"))
	jenkinsSlaveVal := jenkinsSlave

	cb := codebase.Codebase{
		ObjectMeta: metav1.ObjectMeta{
			Name: a.codebaseName,
		},
		Spec: codebase.CodebaseSpec{
			Type:             clusterManagementType,
			Description:      &description,
			DefaultBranch:    defaultBranch,
			Lang:             lang,
			BuildTool:        buildTool,
			Strategy:         "import",
			DeploymentScript: deploymentScript,
			GitServer:        defaultGitServer,
			GitUrlPath:       &repo,
			CiTool:           ciTool,
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				Url: a.repo,
			},
			JenkinsSlave: &jenkinsSlaveVal,
		},
		Status: codebase.CodebaseStatus{
			Available:       false,
			LastTimeUpdated: time.Now(),
			Status:          "initialized",
			Action:          "codebase_registration",
			Value:           "inactive",
		},
	}

	if err := a.codebaseService.Create(&cb); err != nil {
		return nil, errors.Wrap(err, "unable to create cluster codebase")
	}

	if err := a.createTempSecrets(&cb); err != nil {
		return nil, errors.Wrap(err, "unable to create temp secrets")
	}

	if err := a.codebaseService.CreateDefaultBranch(&cb); err != nil {
		return nil, errors.Wrap(err, "unable to create default branch for codebase")
	}

	return &cb, nil
}

func (a *App) createTempSecrets(cb *codebase.Codebase) error {
	secret, err := a.k8sService.GetSecret(a.gerritCreatorSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to get secret")
	}

	username, ok := secret.Data[gerritCreatorUsername]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have username")
	}

	pwd, ok := secret.Data[gerritCreatorPassword]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have password")
	}

	repoSecretName := fmt.Sprintf("repository-codebase-%s-temp", cb.Name)
	if err := a.k8sService.RecreateSecret(repoSecretName, map[string][]byte{
		"username": username,
		"password": pwd,
	}); err != nil {
		return errors.Wrapf(err, "unable to create secret: %s", repoSecretName)
	}

	return nil
}
