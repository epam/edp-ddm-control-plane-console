package controllers

import (
	"context"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/service/logger"
	"ddm-admin-console/util"
	"ddm-admin-console/util/consts"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	projectsV1 "github.com/openshift/api/project/v1"
	"github.com/pkg/errors"
)

var log = logger.GetLogger()

const (
	registryType        = "registry"
	defaultBranch       = "master"
	lang                = "other"
	buildTool           = "gitops"
	strategy            = "clone"
	deploymentScript    = "openshift-template"
	ciTool              = "Jenkins"
	registryOverviewURL = "/admin/registry/overview"
	jenkinsSlave        = "gitops"
)

type CodebaseService interface {
	CreateCodebase(codebase command.CreateCodebase) (*edpv1alpha1.Codebase, error)
	GetCodebaseByName(name string) (*query.Codebase, error)
	UpdateDescription(reg *models.Registry) error
	ExistCodebaseAndBranch(cbName, brName string) bool
	Delete(name, codebaseType string) error
	GetCodebasesByCriteriaK8s(ctx context.Context, criteria query.CodebaseCriteria) ([]*query.Codebase, error)
	GetCodebaseByNameK8s(ctx context.Context, name string) (*query.Codebase, error)
	CreateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error
	UpdateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error
	SetBackupConfig(conf *service.BackupConfig) error
	GetBackupConfig() (*service.BackupConfig, error)
}

type JenkinsService interface {
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
}

type ProjectsService interface {
	GetAll(ctx context.Context) ([]projectsV1.Project, error)
	Get(ctx context.Context, name string) (*projectsV1.Project, error)
}

type EDPComponentServiceK8S interface {
	GetAll(namespace string) ([]v1alpha1.EDPComponent, error)
	Get(namespace, name string) (*v1alpha1.EDPComponent, error)
}

type EditRegistry struct {
	beego.Controller
	BasePath        string
	CodebaseService CodebaseService
	ProjectsService ProjectsService
	JenkinsService  JenkinsService
}

func MakeEditRegistry(codebaseService CodebaseService, projectsService ProjectsService,
	jenkinsService JenkinsService) *EditRegistry {
	return &EditRegistry{
		CodebaseService: codebaseService,
		ProjectsService: projectsService,
		JenkinsService:  jenkinsService,
	}
}

func (r *EditRegistry) Get() {
	r.Data["BasePath"] = r.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	registryName := r.Ctx.Input.Param(":name")
	rg, err := getCodebaseByName(contextWithUserAccessToken(r.Ctx), r.CodebaseService, r.ProjectsService, registryName)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		if _, ok := errors.Cause(err).(service.RegistryNotFound); ok {
			r.TplName = notFoundTemplatePath
			return
		}
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Data["registry"] = rg
}

func (r *EditRegistry) editRegistry(registry *models.Registry) (errorMap map[string][]*validation.Error,
	err error) {
	_, err = getCodebaseByName(contextWithUserAccessToken(r.Ctx), r.CodebaseService, r.ProjectsService, registry.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check registry")
	}

	var valid validation.Validation

	dataValid, err := valid.Valid(registry)
	if err != nil {
		return nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
		return valid.ErrorMap(), nil
	}

	if err = createRegistryKeys(r.CodebaseService, registry, &valid, r.Ctx.Request, false); err != nil {
		err = errors.Wrap(err, "unable to create registry keys")
	}

	if len(valid.ErrorMap()) > 0 {
		return valid.ErrorMap(), nil
	}

	if err := r.CodebaseService.UpdateDescription(registry); err != nil {
		return nil, errors.Wrap(err, "something went wrong during k8s registry edit")
	}

	if err := r.JenkinsService.CreateJobBuildRun(fmt.Sprintf("registry-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", registry.Name, registry.Name), nil); err != nil {
		return nil, errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return
}

func (r *EditRegistry) Post() {
	r.Data["BasePath"] = r.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	var editRegistry models.Registry
	registryName := r.Ctx.Input.Param(":name")
	if err := r.ParseForm(&editRegistry); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}
	editRegistry.Name = registryName

	validationErrors, err := r.editRegistry(&editRegistry)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
		return
	}

	if validationErrors != nil {
		r.Data["registry"] = editRegistry
		log.Error(fmt.Sprintf("%+v\n", validationErrors))
		r.Data["errorsMap"] = validationErrors
		r.Ctx.Output.Status = 422
		if err := r.Render(); err != nil {
			log.Error(err.Error())
		}
		return
	}

	r.Redirect(registryOverviewURL, 303)
}

type ViewRegistry struct {
	beego.Controller
	CodebaseService        CodebaseService
	ProjectsService        ProjectsService
	EDPComponentServiceK8S EDPComponentServiceK8S
	BasePath               string
	Namespace              string
}

func MakeViewRegistry(codebaseService CodebaseService, edpComponentService EDPComponentServiceK8S,
	projectsService ProjectsService, basePath, namespace string) *ViewRegistry {
	return &ViewRegistry{
		CodebaseService:        codebaseService,
		EDPComponentServiceK8S: edpComponentService,
		BasePath:               basePath,
		Namespace:              namespace,
		ProjectsService:        projectsService,
	}
}

func CreateLinksForGerritProviderK8s(edpComponentServiceK8S EDPComponentServiceK8S, registry *query.Codebase,
	namespace string) error {
	cj, err := edpComponentServiceK8S.Get(namespace, consts.Jenkins)
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component resource")
	}

	cg, err := edpComponentServiceK8S.Get(namespace, consts.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit edp component resource")
	}

	for i, b := range registry.CodebaseBranch {
		registry.CodebaseBranch[i].VCSLink = util.CreateGerritLink(cg.Spec.Url, registry.Name, b.Name)
		registry.CodebaseBranch[i].CICDLink = util.CreateCICDApplicationLink(cj.Spec.Url, registry.Name,
			util.ProcessNameToKubernetesConvention(b.Name))
	}

	return nil
}

func (r *ViewRegistry) Get() {
	r.Data["BasePath"] = r.BasePath
	r.Data["Type"] = registryType
	r.TplName = "registry/view.html"
	var gErr error
	defer func() {
		if gErr != nil {
			log.Error(fmt.Sprintf("%+v\n", gErr))
			r.CustomAbort(500, fmt.Sprintf("%+v\n", gErr))
		}
	}()

	registryName := r.Ctx.Input.Param(":name")
	rg, err := getCodebaseByName(contextWithUserAccessToken(r.Ctx), r.CodebaseService, r.ProjectsService, registryName)
	if err != nil {
		if _, ok := errors.Cause(err).(service.RegistryNotFound); ok {
			r.TplName = notFoundTemplatePath
			return
		}

		gErr = err
		return
	}

	if len(rg.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(r.EDPComponentServiceK8S, rg, r.Namespace); err != nil {
			gErr = err
			return
		}
		r.Data["branches"] = rg.CodebaseBranch
	}

	if len(rg.ActionLog) > 0 {
		r.Data["actionLog"] = rg.ActionLog
	}

	ecs, err := r.EDPComponentServiceK8S.GetAll(rg.Name)
	if err != nil {
		gErr = err
		return
	}
	if len(ecs) > 0 {
		r.Data["edpComponents"] = ecs
	}

	r.Data["registry"] = rg
}

//TODO: try to embed this function in all registry controllers
func getCodebaseByName(ctx context.Context, codebaseService CodebaseService, projectsService ProjectsService,
	name string) (*query.Codebase, error) {
	_, err := projectsService.Get(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get project")
	}

	cb, err := codebaseService.GetCodebaseByNameK8s(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebase by name")
	}

	return cb, nil
}
