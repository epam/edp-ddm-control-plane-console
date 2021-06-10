package controllers

import (
	"context"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/util"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	clusterType         = "cluster"
	codebaseDescription = "Керування інфрастуктурними компонентами кластеру"
	defaultGitServer    = "gerrit"
)

type ClusterManagement struct {
	beego.Controller
	CodebaseService      CodebaseService
	JenkinsService       JenkinsService
	EDPComponentsService EDPComponentServiceK8S
	CodebaseName         string
	GitRepo              string
	BasePath             string
	Namespace            string
}

func MakeClusterManagement(jenkinsService JenkinsService, codebaseService CodebaseService,
	edpComponentsService EDPComponentServiceK8S, codebaseName, gitRepo, basePath, namespace string) *ClusterManagement {
	return &ClusterManagement{
		CodebaseService:      codebaseService,
		CodebaseName:         codebaseName,
		EDPComponentsService: edpComponentsService,
		GitRepo:              gitRepo,
		BasePath:             basePath,
		Namespace:            namespace,
		JenkinsService:       jenkinsService,
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

func (c *ClusterManagement) Post() {
	c.Data["BasePath"] = c.BasePath
	c.Data["Type"] = clusterType
	c.TplName = "cluster_management.html"
	c.Data["xsrfdata"] = c.XSRFToken()

	codebase, err := c.CodebaseService.GetCodebaseByNameK8s(context.Background(), c.CodebaseName)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		c.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}
	c.Data["codebase"] = codebase

	if len(codebase.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(c.EDPComponentsService, codebase, c.Namespace); err != nil {
			log.Error(fmt.Sprintf("%+v\n", err))
			c.CustomAbort(500, fmt.Sprintf("%+v\n", err))
			return
		}
		c.Data["branches"] = codebase.CodebaseBranch
	}

	backupConfig, validationErrors, err := c.setBackupConfig()
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		c.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	if validationErrors != nil {
		c.Data["errorsMap"] = validationErrors
	}

	c.Data["backupConf"] = backupConfig
}

func (c *ClusterManagement) setBackupConfig() (backupConfig *service.BackupConfig, errorMap map[string][]*validation.Error, err error) {
	backupConfig = &service.BackupConfig{}
	if err := c.ParseForm(backupConfig); err != nil {
		return nil, nil, errors.Wrap(err, "unable to parse form")
	}

	var valid validation.Validation
	dataValid, err := valid.Valid(backupConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
		return backupConfig, valid.ErrorMap(), nil
	}

	if err := c.CodebaseService.SetBackupConfig(backupConfig); err != nil {
		return nil, nil, errors.Wrap(err, "unable to set backup config")
	}

	if err := c.JenkinsService.CreateJobBuildRun(fmt.Sprintf("cluster-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", c.CodebaseName, c.CodebaseName), nil); err != nil {
		return nil, nil, errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return backupConfig, nil, nil
}

func (c *ClusterManagement) Get() {
	c.Data["xsrfdata"] = c.XSRFToken()
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

	codebase, err := c.CodebaseService.GetCodebaseByNameK8s(context.Background(), c.CodebaseName)
	if err != nil {
		if service.IsRegistryNotFound(err) {
			codebase, err = c.createClusterCodebase()
			if err != nil {
				gErr = err
				return
			}

			return
		}

		gErr = err
		return
	}

	conf, err := c.CodebaseService.GetBackupConfig()
	if err != nil && !k8sErrors.IsNotFound(errors.Cause(err)) {
		gErr = err
		return
	}
	if conf == nil {
		conf = &service.BackupConfig{}
	}
	c.Data["backupConf"] = conf

	if len(codebase.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(c.EDPComponentsService, codebase, c.Namespace); err != nil {
			gErr = err
			return
		}
		c.Data["branches"] = codebase.CodebaseBranch
	}

	c.Data["codebase"] = codebase
}
