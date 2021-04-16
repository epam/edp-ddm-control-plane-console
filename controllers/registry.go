package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/util"
	"ddm-admin-console/util/consts"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/pkg/errors"
)

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
	GetCodebasesByCriteria(criteria query.CodebaseCriteria) ([]*query.Codebase, error)
	GetCodebaseByName(name string) (*query.Codebase, error)
	UpdateDescription(reg *models.Registry) error
	ExistCodebaseAndBranch(cbName, brName string) bool
	Delete(name, codebaseType string) error
	GetCodebasesByCriteriaK8s(criteria query.CodebaseCriteria) ([]*query.Codebase, error)
	GetCodebaseByNameK8s(name string) (*query.Codebase, error)
	CreateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error
}

type EDPComponentServiceK8S interface {
	GetAll(namespace string) ([]v1alpha1.EDPComponent, error)
	Get(namespace, name string) (*v1alpha1.EDPComponent, error)
}

type ListRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
}

func MakeListRegistry(codebaseService CodebaseService) *ListRegistry {
	return &ListRegistry{
		CodebaseService: codebaseService,
	}
}

func (r *ListRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["Type"] = registryType
	r.Data["xsrfdata"] = r.XSRFToken()

	r.TplName = "registry/list.html"

	codebases, err := r.CodebaseService.GetCodebasesByCriteriaK8s(query.CodebaseCriteria{
		Type: query.Registry,
	})

	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Data["registries"] = codebases
}

func (r *ListRegistry) Post() {
	registryName := r.GetString("registry-name")
	rg, err := r.CodebaseService.GetCodebaseByNameK8s(registryName)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		if _, ok := errors.Cause(err).(service.RegistryNotFound); ok {
			r.TplName = notFoundTemplatePath
			return
		}

		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	if err := r.CodebaseService.Delete(rg.Name, string(rg.Type)); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Redirect(registryOverviewURL, 303)
}

type EditRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
}

func MakeEditRegistry(codebaseService CodebaseService) *EditRegistry {
	return &EditRegistry{
		CodebaseService: codebaseService,
	}
}

func (r *EditRegistry) Get() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	registryName := r.Ctx.Input.Param(":name")
	rg, err := r.CodebaseService.GetCodebaseByNameK8s(registryName)
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
	var valid validation.Validation

	dataValid, err := valid.Valid(registry)
	if err != nil {
		return nil, errors.Wrap(err, "something went wrong during validation")
	}

	if !dataValid {
		return valid.ErrorMap(), nil
	}

	if err := r.CodebaseService.UpdateDescription(registry); err != nil {
		return nil, errors.Wrap(err, "something went wrong during k8s registry edit")
	}

	return
}

func (r *EditRegistry) Post() {
	r.Data["BasePath"] = console.BasePath
	r.Data["xsrfdata"] = r.XSRFToken()
	r.Data["Type"] = registryType
	r.TplName = "registry/edit.html"

	var parsedRegistry models.Registry
	registryName := r.Ctx.Input.Param(":name")
	if err := r.ParseForm(&parsedRegistry); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}
	parsedRegistry.Name = registryName

	validationErrors, err := r.editRegistry(&parsedRegistry)
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, err.Error())
		return
	}

	if validationErrors != nil {
		r.Data["registry"] = parsedRegistry
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
	EDPComponentServiceK8S EDPComponentServiceK8S
}

func MakeViewRegistry(codebaseService CodebaseService, edpComponentService EDPComponentServiceK8S) *ViewRegistry {
	return &ViewRegistry{
		CodebaseService:        codebaseService,
		EDPComponentServiceK8S: edpComponentService,
	}
}

func CreateLinksForGerritProviderK8s(edpComponentServiceK8S EDPComponentServiceK8S, registry *query.Codebase) error {
	cj, err := edpComponentServiceK8S.Get(console.Namespace, consts.Jenkins)
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component resource")
	}

	cg, err := edpComponentServiceK8S.Get(console.Namespace, consts.Gerrit)
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
	r.Data["BasePath"] = console.BasePath
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
	rg, err := r.CodebaseService.GetCodebaseByNameK8s(registryName)
	if err != nil {
		if _, ok := errors.Cause(err).(service.RegistryNotFound); ok {
			r.TplName = notFoundTemplatePath
			return
		}

		gErr = err
		return
	}

	if len(rg.CodebaseBranch) > 0 {
		if err := CreateLinksForGerritProviderK8s(r.EDPComponentServiceK8S, rg); err != nil {
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
