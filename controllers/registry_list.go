package controllers

import (
	"context"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"fmt"

	"github.com/astaxie/beego"
	projectsV1 "github.com/openshift/api/project/v1"
	"github.com/pkg/errors"
)

type ListRegistry struct {
	beego.Controller
	CodebaseService CodebaseService
	ProjectsService ProjectsService
	BasePath        string
}

func MakeListRegistry(codebaseService CodebaseService, projectsService ProjectsService) *ListRegistry {
	return &ListRegistry{
		CodebaseService: codebaseService,
		ProjectsService: projectsService,
	}
}

func (r *ListRegistry) Get() {
	r.Data["BasePath"] = r.BasePath
	r.Data["Type"] = registryType
	r.Data["xsrfdata"] = r.XSRFToken()

	r.TplName = "registry/list.html"

	codebases, err := r.getUserCodebases()
	if err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Data["registries"] = codebases
}

func (r *ListRegistry) getUserCodebases() ([]*query.Codebase, error) {
	codebases, err := r.CodebaseService.GetCodebasesByCriteriaK8s(context.Background(),
		query.CodebaseCriteria{
			Type: query.Registry,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list codebases")
	}

	projects, err := r.ProjectsService.GetAll(contextWithUserAccessToken(r.Ctx))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user projects")
	}

	return filterCodebasesByAvailableProjects(codebases, projects), nil
}

func filterCodebasesByAvailableProjects(codebases []*query.Codebase, projects []projectsV1.Project) []*query.Codebase {
	prjs := make(map[string]struct{})
	for _, p := range projects {
		prjs[p.Name] = struct{}{}
	}

	filteredCodebases := make([]*query.Codebase, 0, len(codebases))
	for _, cb := range codebases {
		if _, ok := prjs[cb.Name]; ok {
			filteredCodebases = append(filteredCodebases, cb)
		}
	}

	return filteredCodebases
}

func (r *ListRegistry) Post() {
	registryName := r.GetString("registry-name")
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

	if err := r.CodebaseService.Delete(rg.Name, string(rg.Type)); err != nil {
		log.Error(fmt.Sprintf("%+v\n", err))
		r.CustomAbort(500, fmt.Sprintf("%+v\n", err))
		return
	}

	r.Redirect(registryOverviewURL, 303)
}
