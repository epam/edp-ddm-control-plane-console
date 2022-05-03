package cluster

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
)

func (a *App) view(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.k8sService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s service for user")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.codebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}

	cbService, err := a.codebaseService.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init codebase service for user")
	}

	cb, err := cbService.Get(a.codebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster codebase")
	}

	branches, err := cbService.GetBranchesByCodebase(cb.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry branches")
	}
	cb.Branches = branches

	jenkinsComponent, err := a.edpComponentService.Get(userCtx, "jenkins")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.edpComponentService.Get(userCtx, "gerrit")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.edpComponentService.GetAllNamespace(userCtx, cb.Name, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaced edp components")
	}

	mrs, err := a.gerritService.GetMergeRequestByProject(ctx, a.codebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit merge requests")
	}

	return router.MakeResponse(200, "cluster/view.html", gin.H{
		"branches":         branches,
		"codebase":         cb,
		"jenkinsURL":       jenkinsComponent.Spec.Url,
		"gerritURL":        gerritComponent.Spec.Url,
		"page":             "cluster",
		"edpComponents":    namespacedEDPComponents,
		"canUpdateCluster": canUpdateCluster,
		"mergeRequests":    mrs,
	}), nil
}
