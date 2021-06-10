package controllers

import (
	"ddm-admin-console/service"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

type ClusterEdit struct {
	beego.Controller
	JenkinsService  JenkinsService
	CodebaseService CodebaseService
	Namespace       string
	CodebaseName    string
}

func MakeClusterEdit(jenkinsService JenkinsService, codebaseService CodebaseService, namespace, codebaseName string) *ClusterEdit {
	return &ClusterEdit{
		Namespace:       namespace,
		CodebaseService: codebaseService,
		JenkinsService:  jenkinsService,
		CodebaseName:    codebaseName,
	}
}

func (c *ClusterEdit) Get() {
	c.Data["xsrfdata"] = c.XSRFToken()
	c.Data["Type"] = clusterType
	c.TplName = "cluster/edit.html"

	var gErr error
	defer func() {
		if gErr != nil {
			log.Error(fmt.Sprintf("%+v\n", gErr))
			c.CustomAbort(500, fmt.Sprintf("%+v\n", gErr))
		}
	}()

	conf, err := c.CodebaseService.GetBackupConfig()
	if err != nil && !k8sErrors.IsNotFound(errors.Cause(err)) {
		gErr = err
		return
	}
	if conf == nil {
		conf = &service.BackupConfig{}
	}
	c.Data["backupConf"] = conf
}

func (c *ClusterEdit) Post() {
	c.Data["xsrfdata"] = c.XSRFToken()
	c.Data["Type"] = clusterType
	c.TplName = "cluster/edit.html"

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
	c.Redirect("/admin/cluster/management", 302)
}

func (c *ClusterEdit) setBackupConfig() (backupConfig *service.BackupConfig, errorMap map[string][]*validation.Error, err error) {
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
