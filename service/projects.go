package service

import (
	"context"
	"ddm-admin-console/k8s"

	projectsV1 "github.com/openshift/api/project/v1"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Projects struct {
	clients *k8s.ClientSet
}

func MakeProjects(clients *k8s.ClientSet) *Projects {
	return &Projects{
		clients: clients,
	}
}

func (p *Projects) GetAll(ctx context.Context) ([]projectsV1.Project, error) {
	pc, err := p.clients.GetOCProjectsClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get oc projects client")
	}

	prjs, err := pc.Projects().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list projects")
	}

	return prjs.Items, nil
}

func (p *Projects) Get(ctx context.Context, name string) (*projectsV1.Project, error) {
	pc, err := p.clients.GetOCProjectsClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get oc projects client")
	}

	prj, err := pc.Projects().Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get project")
	}

	return prj, nil
}
