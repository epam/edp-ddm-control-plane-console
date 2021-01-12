package controllers

import (
	"ddm-admin-console/console"
	validation2 "ddm-admin-console/controllers/validation"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	cbs "ddm-admin-console/service/codebasebranch"
	"ddm-admin-console/util"
	"ddm-admin-console/util/auth"
	"ddm-admin-console/util/consts"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"go.uber.org/zap"
)

const redirectURL = `%s/admin/autotest/create`

type AutotestsController struct {
	beego.Controller
	CodebaseService  service.CodebaseService
	EDPTenantService service.EDPTenantService
	BranchService    cbs.Service
	GitServerService service.GitServerService
	SlaveService     service.SlaveService
	JobProvisioning  service.JobProvisioning

	IntegrationStrategies []string
	BuildTools            []string
	VersioningTypes       []string
	TestReportTools       []string
	DeploymentScript      []string
	CiTools               []string
}

func (c *AutotestsController) CreateAutotests() {
	flash := beego.NewFlash()
	codebase := c.extractAutotestsRequestData()
	errMsg := validateAutotestsRequestData(codebase)
	if errMsg != nil {
		log.Error("Failed to validate autotests request data", zap.String("err", errMsg.Message))
		flash := beego.NewFlash()
		flash.Error(errMsg.Message)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf(redirectURL, console.BasePath), 302)
		return
	}
	logAutotestsRequestData(codebase)

	createdObject, err := c.CodebaseService.CreateCodebase(codebase)
	if err != nil {
		c.checkError(err, flash, codebase.Name, codebase.GitURLPath)
		return
	}

	log.Info("Autotests object is saved into cluster", zap.String("name", createdObject.Name))
	flash.Success("Autotests object is created.")
	flash.Store(&c.Controller)
	c.Redirect(fmt.Sprintf("%s/admin/autotest/overview?%s=%s#codebaseSuccessModal", console.BasePath, paramWaitingForCodebase, codebase.Name), 302)
}

func (c *AutotestsController) checkError(err error, flash *beego.FlashData, name string, url *string) {
	switch err.(type) {
	case *edperror.CodebaseAlreadyExistsError:
		flash.Error("Autotest %v already exists.", name)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf(redirectURL, console.BasePath), 302)
	case *edperror.CodebaseWithGitURLPathAlreadyExistsError:
		flash.Error("Autotest %v with %v project path already exists.", name, *url)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf(redirectURL, console.BasePath), 302)
	default:
		log.Error("couldn't create codebase", zap.Error(err))
		c.Abort("500")
	}
}

func logAutotestsRequestData(autotests command.CreateCodebase) {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Request data to create codebase CR is valid. name=%s, strategy=%s, lang=%s, buildTool=%s, testReportFramework=%s",
		autotests.Name, autotests.Strategy, autotests.Lang, autotests.BuildTool, *autotests.TestReportFramework))

	if autotests.Repository != nil {
		result.WriteString(fmt.Sprintf(", repositoryUrl=%s, repositoryLogin=%s", autotests.Repository.URL, autotests.Repository.Login))
	}

	if autotests.Vcs != nil {
		result.WriteString(fmt.Sprintf(", vcsLogin=%s", autotests.Vcs.Login))
	}

	log.Info(result.String())
}

func (c *AutotestsController) extractAutotestsRequestData() command.CreateCodebase {
	codebase := command.CreateCodebase{
		Lang:             c.GetString("appLang"),
		BuildTool:        c.GetString("buildTool"),
		Strategy:         strings.ToLower(c.GetString("strategy")),
		Type:             "autotests",
		DeploymentScript: c.GetString("deploymentScript"),
		Name:             c.GetString("appName"),
		CiTool:           c.GetString("ciTool"),
		DefaultBranch:    c.GetString("defaultBranchName"),
	}

	if js := c.GetString("jenkinsSlave"); len(js) > 0 {
		codebase.JenkinsSlave = &js
	}

	if jp := c.GetString("jobProvisioning"); len(jp) > 0 {
		codebase.JobProvisioning = &jp
	}

	if s := c.GetString("jiraServer"); len(s) > 0 {
		codebase.JiraServer = &s
	}

	if v := c.GetString("commitMessagePattern"); len(v) > 0 {
		codebase.CommitMessageRegex = &v
	}

	if v := c.GetString("ticketNamePattern"); len(v) > 0 {
		codebase.TicketNameRegex = &v
	}

	codebase.Versioning.Type = c.GetString("versioningType")
	startVersioningFrom := c.GetString("startVersioningFrom")
	sp := c.GetString("snapshotStaticField")
	codebase.Versioning.StartFrom = util.GetVersionOrNil(startVersioningFrom, sp)

	framework := c.GetString("framework")
	codebase.Framework = &framework

	if codebase.Strategy == consts.ImportStrategy {
		codebase.GitServer = c.GetString("gitServer")
		gitRepoPath := c.GetString("gitRelativePath")
		codebase.GitURLPath = &gitRepoPath
	} else {
		codebase.GitServer = defaultGitServer
	}

	testReportFramework := c.GetString("testReportFramework")
	if testReportFramework != "" {
		codebase.TestReportFramework = &testReportFramework
	}

	repoURL := c.GetString("gitRepoUrl")
	if repoURL != "" {
		codebase.Repository = &command.Repository{
			URL: repoURL,
		}

		isRepoPrivate, _ := c.GetBool("isRepoPrivate", false)
		if isRepoPrivate {
			codebase.Repository.Login = c.GetString("repoLogin")
			codebase.Repository.Password = c.GetString("repoPassword")
		}
	}

	description := c.GetString("description")
	if description != "" {
		codebase.Description = &description
	}

	vcsLogin := c.GetString("vcsLogin")
	vcsPassword := c.GetString("vcsPassword")
	if vcsLogin != "" && vcsPassword != "" {
		codebase.Vcs = &command.Vcs{
			Login:    vcsLogin,
			Password: vcsPassword,
		}
	}
	codebase.Username, _ = c.Ctx.Input.Session("username").(string)
	return codebase
}

func validateAutotestsRequestData(autotests command.CreateCodebase) *validation2.ErrMsg {
	valid := validation.Validation{}

	_, err := valid.Valid(autotests)

	if autotests.Strategy == consts.ImportStrategy {
		valid.Match(autotests.GitURLPath, regexp.MustCompile("^\\/.*$"), "Spec.GitURLPath") //nolint
	}

	if autotests.Repository != nil {
		_, err = valid.Valid(autotests.Repository)

		isAvailable := util.IsGitRepoAvailable(autotests.Repository.URL, autotests.Repository.Login, autotests.Repository.Password)

		if !isAvailable {
			err := &validation.Error{Key: "repository", Message: "Repository doesn't exist or invalid login and password."}
			valid.Errors = append(valid.Errors, err)
		}
	}

	if autotests.Description == nil {
		err := &validation.Error{Key: "description", Message: "Description field can't be empty."}
		valid.Errors = append(valid.Errors, err)
	}

	if autotests.Vcs != nil {
		_, err = valid.Valid(autotests.Vcs)
	}

	if err != nil {
		return &validation2.ErrMsg{
			Message:    "An internal error has occurred on server while validating autotests's form fields.",
			StatusCode: http.StatusInternalServerError}
	}

	if valid.Errors == nil {
		return nil
	}

	return &validation2.ErrMsg{Message: string(validation2.CreateErrorResponseBody(valid)),
		StatusCode: http.StatusBadRequest}
}

func (c *AutotestsController) GetCreateAutotestsPage() {
	flash := beego.ReadFromRequest(&c.Controller)
	if flash.Data["error"] != "" {
		c.Data["Error"] = flash.Data["error"]
	}

	isVcsEnabled, err := c.EDPTenantService.GetVcsIntegrationValue()
	if err != nil {
		c.Abort("500")
		return
	}

	contains := doesIntegrationStrategiesContainImportStrategy(c.IntegrationStrategies)
	if contains {
		log.Info("Import strategy is used.")

		gitServers, err := c.GitServerService.GetServers(query.GitServerCriteria{Available: true})
		if err != nil {
			c.Abort("500")
			return
		}
		log.Info("Fetched Git Servers", zap.Any("git servers", gitServers))

		c.Data["GitServers"] = gitServers
	}

	s, err := c.SlaveService.GetAllSlaves()
	if err != nil {
		c.Abort("500")
		return
	}

	p, err := c.JobProvisioning.GetAllJobProvisioners(query.JobProvisioningCriteria{Scope: util.GetStringP(scope)})
	if err != nil {
		c.Abort("500")
		return
	}

	log.Info("Create strategy is removed from list due to Autotest")

	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["HasRights"] = auth.IsAdmin(c.GetSession("realm_roles").([]string))
	c.Data["IsVcsEnabled"] = isVcsEnabled
	c.Data["Type"] = query.Autotests
	c.Data["IntegrationStrategies"] = c.IntegrationStrategies
	c.Data["CodeBaseIntegrationStrategy"] = true
	c.Data["JenkinsSlaves"] = s
	c.Data["BuildTools"] = c.BuildTools
	c.Data["TestReportTools"] = c.TestReportTools
	c.Data["JobProvisioners"] = p
	c.Data["VersioningTypes"] = c.VersioningTypes
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["BasePath"] = console.BasePath
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.Data["CiTools"] = c.CiTools
	c.TplName = "create_autotest.html"
}

func (c *AutotestsController) GetAutotestsOverviewPage() {
	flash := beego.ReadFromRequest(&c.Controller)
	codebases, err := c.CodebaseService.GetCodebasesByCriteria(query.CodebaseCriteria{
		Type: query.Autotests,
	})
	codebases = addCodebaseInProgressIfAny(codebases, c.GetString(paramWaitingForCodebase))
	if err != nil {
		c.Abort("500")
		return
	}

	if flash.Data["success"] != "" {
		c.Data["Success"] = true
	}
	if flash.Data["error"] != "" {
		c.Data["DeletionError"] = flash.Data["error"]
	}
	c.Data["Codebases"] = codebases
	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["HasRights"] = auth.IsAdmin(c.GetSession("realm_roles").([]string))
	c.Data["Type"] = query.Autotests
	c.Data["BasePath"] = console.BasePath
	c.Data["VersioningTypes"] = c.VersioningTypes
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.TplName = tplCodebase
}
