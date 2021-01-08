/*
 * Copyright 2020 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/controllers/validation"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	cbs "ddm-admin-console/service/codebasebranch"
	"ddm-admin-console/util"
	"ddm-admin-console/util/auth"
	"ddm-admin-console/util/consts"
	dberror "ddm-admin-console/util/error/db-errors"
	"fmt"
	"html/template"

	"github.com/astaxie/beego"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type EDPComponentService interface {
	GetEDPComponent(componentType string) (*query.EDPComponent, error)
}

type CodebaseController struct {
	beego.Controller
	CodebaseService  service.CodebaseService
	EDPTenantService service.EDPTenantService
	BranchService    cbs.Service
	GitServerService service.GitServerService
	EDPComponent     EDPComponentService
}

const (
	paramWaitingForBranch = "waitingforbranch"
	codebaseTypeAutotest  = "autotest"
	defaultGitServer      = "gerrit"
)

func (c *CodebaseController) GetCodebaseOverviewPage() {
	cn := c.GetString(":codebaseName")
	log.Debug("start GetCodebaseOverviewPage method from controller", zap.String("name", cn))
	codebase, err := c.CodebaseService.GetCodebaseByName(cn)
	if err != nil {
		log.Error(err.Error())
		c.Abort("500")
		return
	}

	if codebase == nil {
		log.Error("codebase wasn't found", zap.String("name", cn))
		c.Abort("404")
		return
	}

	if err := c.createBranchLinks(codebase, console.Tenant); err != nil {
		log.Error("an error has occurred while creating link to Git provider", zap.Error(err))
		c.Abort("500")
		return
	}

	codebase.CodebaseBranch = addCodebaseBranchInProgressIfAny(codebase.CodebaseBranch, c.GetString(paramWaitingForBranch))

	flash := beego.ReadFromRequest(&c.Controller)
	if flash.Data["error"] != "" {
		c.Data["ErrorBranch"] = flash.Data["error"]
	}

	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["Codebase"] = codebase
	c.Data["Type"] = codebase.Type
	c.Data["HasRights"] = auth.IsAdmin(c.GetSession("realm_roles").([]string))
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["BasePath"] = console.BasePath
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.setCodebaseTypeCaptions(string(codebase.Type))
	c.TplName = "codebase_overview.html"
}

func (c *CodebaseController) setCodebaseTypeCaptions(codebaseType string) {
	switch codebaseType {
	case "application":
		{
			c.Data["TypeCaption"] = "Application"
			c.Data["TypeSingular"] = "application"
		}
	case "autotests":
		{
			c.Data["TypeCaption"] = "Autotests codebase"
			c.Data["TypeSingular"] = "autotests codebase"
		}
	case "library":
		{
			c.Data["TypeCaption"] = "Library"
			c.Data["TypeSingular"] = "library"
		}
	}
}

func addCodebaseBranchInProgressIfAny(branches []*query.CodebaseBranch, branchInProgress string) []*query.CodebaseBranch {
	if branchInProgress != "" {
		for _, branch := range branches {
			if branch.Name == branchInProgress {
				return branches
			}
		}

		log.Debug("Adding branch which is going to be created to the list.",
			zap.String("name", branchInProgress))
		branch := query.CodebaseBranch{
			Name:   branchInProgress,
			Status: "inactive",
		}
		branches = append(branches, &branch)
	}
	return branches
}

func (c CodebaseController) createBranchLinks(codebase *query.Codebase, tenant string) error {
	if codebase.Strategy == consts.ImportStrategy {
		return c.createLinksForGitProvider(codebase)
	}

	return CreateLinksForGerritProvider(c.EDPComponent, codebase)
}

func (c CodebaseController) createLinksForGitProvider(codebase *query.Codebase) error {
	g, err := c.GitServerService.GetGitServer(*codebase.GitServer)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New(fmt.Sprintf("unexpected behaviour. couldn't find %v GitServer in DB", *codebase.GitServer)) //nolint
	}

	jc, err := c.EDPComponent.GetEDPComponent(consts.Jenkins)
	if err != nil {
		return err
	}

	if jc == nil {
		return fmt.Errorf("jenkin link can't be created for %v codebase because of edp-component %v is absent in DB",
			codebase.Name, consts.Jenkins)
	}

	for i, b := range codebase.CodebaseBranch {
		codebase.CodebaseBranch[i].VCSLink = util.CreateGitLink(g.Hostname, *codebase.GitProjectPath, b.Name)
		codebase.CodebaseBranch[i].CICDLink = getCiLink(codebase, jc.URL, b.Name, g.Hostname)
	}

	return nil
}

func getCiLink(codebase *query.Codebase, jenkinsHost, branch, gitHost string) string {
	if consts.JenkinsCITool == codebase.CiTool {
		return util.CreateCICDApplicationLink(jenkinsHost, codebase.Name, util.ProcessNameToKubernetesConvention(branch))
	}
	return util.CreateGitlabCILink(gitHost, *codebase.GitProjectPath)
}

func CreateLinksForGerritProvider(edpComponent EDPComponentService, codebase *query.Codebase) error {
	cj, err := edpComponent.GetEDPComponent(consts.Jenkins)
	if err != nil {
		return errors.Wrap(err, "unable to get Jenkins edp component")
	}

	if cj == nil {
		return fmt.Errorf("jenkin link can't be created for %v codebase because of edp-component %v is absent in DB",
			codebase.Name, consts.Jenkins)
	}

	cg, err := edpComponent.GetEDPComponent(consts.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get Gerrit edp component")
	}

	if cg == nil {
		return fmt.Errorf("gerrit link can't be created for %v codebase because of edp-component %v is absent in DB",
			codebase.Name, consts.Gerrit)
	}

	for i, b := range codebase.CodebaseBranch {
		codebase.CodebaseBranch[i].VCSLink = util.CreateGerritLink(cg.URL, codebase.Name, b.Name)
		codebase.CodebaseBranch[i].CICDLink = util.CreateCICDApplicationLink(cj.URL, codebase.Name,
			util.ProcessNameToKubernetesConvention(b.Name))
	}

	return nil
}

func (c *CodebaseController) Delete() {
	flash := beego.NewFlash()
	cn := c.GetString("name")
	log.Debug("delete codebase method is invoked", zap.String("name", cn))
	ct := c.GetString("codebase-type")
	if err := c.CodebaseService.Delete(cn, ct); err != nil {
		if dberror.CodebaseIsUsed(err) {
			cerr, _ := err.(dberror.CodebaseIsUsedByCDPipeline)
			flash.Error(cerr.Message)
			flash.Store(&c.Controller)
			log.Error(cerr.Message, zap.Error(err))
			c.Redirect(createCodebaseIsUsedURL(cerr.Codebase, ct), 302)
			return
		}
		log.Error("delete process is failed", zap.Error(err))
		c.Abort("500")
		return
	}
	log.Info("delete codebase method is finished", zap.String("name", cn))
	c.Redirect(createCodebaseIsDeletedURL(cn, ct), 302)
}

func createCodebaseIsUsedURL(codebaseName, codebaseType string) string {
	if codebaseType == consts.Autotest {
		codebaseType = codebaseTypeAutotest
	}
	return fmt.Sprintf("%s/admin/%v/overview?codebase=%v#codebaseIsUsed", console.BasePath,
		codebaseType, codebaseName)
}

func createCodebaseIsDeletedURL(codebaseName, codebaseType string) string {
	if codebaseType == consts.Autotest {
		codebaseType = codebaseTypeAutotest
	}
	return fmt.Sprintf("%s/admin/%v/overview?codebase=%v#codebaseIsDeleted", console.BasePath, codebaseType, codebaseName)
}

func (c *CodebaseController) GetEditCodebasePage() {
	flash := beego.ReadFromRequest(&c.Controller)
	if flash.Data["error"] != "" {
		c.Data["CodebaseUpdateError"] = flash.Data["error"]
	}

	n := c.GetString(":name")
	log.Debug("start executing GetEditCodebasePage method", zap.String("name", n))

	codebase, err := c.CodebaseService.GetCodebaseByName(n)
	if err != nil {
		log.Error("couldn't get codebase from db", zap.Error(err))
		c.Abort("500")
		return
	}

	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["BasePath"] = console.BasePath
	c.Data["Codebase"] = codebase
	c.Data["Type"] = codebase.Type
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.TplName = "edit_codebase.html"
}

func (c *CodebaseController) Update() {
	flash := beego.NewFlash()
	name := c.GetString("name")
	log.Debug("start executing Update method", zap.String("name", name))

	cc := command.UpdateCodebaseCommand{
		Name:               name,
		CommitMessageRegex: c.GetString("commitMessagePattern"),
		TicketNameRegex:    c.GetString("ticketNamePattern"),
	}

	errMsg := validation.ValidateCodebaseUpdateRequestData(cc)
	if errMsg != nil {
		log.Error("Codebase update request data is invalid", zap.String("err", errMsg.Message))
		flash.Error(errMsg.Message)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf("%v/admin/codebase/%v/update", console.BasePath, cc.Name), 302)
		return
	}

	codebase, err := c.CodebaseService.Update(cc)
	if err != nil {
		log.Error("couldn't update codebase", zap.Error(err))
		c.Abort("500")
		return
	}

	c.Redirect(fmt.Sprintf("%v/admin/%v/overview#codebaseUpdateSuccessModal",
		console.BasePath, getType(codebase.Spec.Type)), 302)
}

func getType(codebaseType string) string {
	if codebaseType == "autotests" {
		return codebaseTypeAutotest
	}
	return codebaseType
}
