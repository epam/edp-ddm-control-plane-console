package controllers

import (
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/util"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/astaxie/beego"
)

const (
	clusterType         = "cluster"
	codebaseDescription = "Керування інфрастуктурними компонентами кластеру"
	defaultGitServer    = "gerrit"
)

type ClusterManagement struct {
	beego.Controller
	CodebaseService      CodebaseService
	EDPComponentsService EDPComponentServiceK8S
	CodebaseName         string
	GitRepo              string
	BasePath             string
	Namespace            string
}

func MakeClusterManagement(codebaseService CodebaseService, edpComponentsService EDPComponentServiceK8S,
	codebaseName, gitRepo, basePath, namespace string) *ClusterManagement {
	return &ClusterManagement{
		CodebaseService:      codebaseService,
		CodebaseName:         codebaseName,
		EDPComponentsService: edpComponentsService,
		GitRepo:              gitRepo,
		BasePath:             basePath,
		Namespace:            namespace,
	}
}

func (c *ClusterManagement) createClusterCodebase() (*query.Codebase, error) {
	username, _ := c.Ctx.Input.Session("username").(string)
	jobProvisioning := "default"
	startVersion := "0.0.1"
	versioning := command.Versioning{
		StartFrom: &startVersion,
		Type:      "edp",
	}

	if !strings.Contains(c.GitRepo, "//") || !strings.Contains(c.GitRepo, "/") {
		return nil, errors.New("wrong git repo")
	}

	repoPath := strings.Join(strings.Split(strings.Split(c.GitRepo, "//")[1], "/")[1:], "/")

	_, err := c.CodebaseService.CreateCodebase(command.CreateCodebase{
		Name:             c.CodebaseName,
		Username:         username,
		Type:             string(query.ClusterManagement),
		Description:      util.GetStringP(codebaseDescription),
		DefaultBranch:    defaultBranch,
		Lang:             lang,
		BuildTool:        buildTool,
		Strategy:         "import",
		DeploymentScript: deploymentScript,
		GitServer:        defaultGitServer,
		GitURLPath:       util.GetStringP(fmt.Sprintf("/%s", repoPath)),
		CiTool:           ciTool,
		JobProvisioning:  &jobProvisioning,
		Versioning:       versioning,
		Repository: &command.Repository{
			URL: c.GitRepo,
		},
		JenkinsSlave: util.GetStringP(jenkinsSlave),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create cluster codebase")
	}

	codebase, err := c.CodebaseService.GetCodebaseByNameK8s(contextWithUserAccessToken(c.Ctx), c.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load cluster codebase")
	}

	return codebase, nil
}

func (c *ClusterManagement) Get() {
	c.Data["BasePath"] = c.BasePath
	c.Data["Type"] = clusterType
	c.TplName = "cluster_management.html"
	var gErr error
	defer func() {
		if gErr != nil {
			log.Error(fmt.Sprintf("%+v\n", gErr))
			c.CustomAbort(500, fmt.Sprintf("%+v\n", gErr))
		}
	}()

	codebase, err := c.CodebaseService.GetCodebaseByNameK8s(contextWithUserAccessToken(c.Ctx), c.CodebaseName)
	if err != nil {
		switch err.(type) {
		case service.RegistryNotFound:
			codebase, err = c.createClusterCodebase()
			if err != nil {
				gErr = err
				return
			}
		default:
			gErr = err
			return
		}
	}

	if len(codebase.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(c.EDPComponentsService, codebase, c.Namespace); err != nil {
			gErr = err
			return
		}
		c.Data["branches"] = codebase.CodebaseBranch
	}

	c.Data["codebase"] = codebase
}
