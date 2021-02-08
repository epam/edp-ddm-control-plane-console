package controllers

import (
	"ddm-admin-console/console"
	"fmt"

	"github.com/astaxie/beego"
)

const (
	clusterType = "cluster"
)

type ClusterManagement struct {
	beego.Controller
	CodebaseService      CodebaseService
	EDPComponentsService EDPComponentServiceK8S
	CodebaseName         string
}

func MakeClusterManagement(codebaseService CodebaseService, edpComponentsService EDPComponentServiceK8S,
	codebaseName string) *ClusterManagement {
	return &ClusterManagement{
		CodebaseService:      codebaseService,
		CodebaseName:         codebaseName,
		EDPComponentsService: edpComponentsService,
	}
}

func (c *ClusterManagement) Get() {
	c.Data["BasePath"] = console.BasePath
	c.Data["Type"] = clusterType
	c.TplName = "cluster_management.html"
	var gErr error
	defer func() {
		if gErr != nil {
			log.Error(fmt.Sprintf("%+v\n", gErr))
			c.CustomAbort(500, fmt.Sprintf("%+v\n", gErr))
		}
	}()

	codebase, err := c.CodebaseService.GetCodebaseByNameK8s(c.CodebaseName)
	if err != nil {
		gErr = err
		return
	}

	if len(codebase.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(c.EDPComponentsService, codebase); err != nil {
			gErr = err
			return
		}
		c.Data["branches"] = codebase.CodebaseBranch
	}

	c.Data["codebase"] = codebase
}
