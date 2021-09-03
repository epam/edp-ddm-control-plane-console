package gerrit

import (
	"context"
	"ddm-admin-console/service"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

type Service struct {
	service.UserConfig
	k8sClient  client.Client
	scheme     *runtime.Scheme
	namespace  string
	restConfig *rest.Config
}

func Make(k8sConfig *rest.Config, namespace string) (*Service, error) {
	s := runtime.NewScheme()
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&GerritProject{}, &GerritProjectList{})

	if err := builder.AddToScheme(s); err != nil {
		return nil, errors.Wrap(err, "error during builder add to scheme")
	}

	cl, err := client.New(k8sConfig, client.Options{
		Scheme: s,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s jenkins client")
	}

	return &Service{
		k8sClient: cl,
		scheme:    s,
		namespace: namespace,
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		restConfig: k8sConfig,
	}, nil
}

func (s *Service) GetProjects(ctx context.Context) ([]GerritProject, error) {
	var projectList GerritProjectList
	if err := s.k8sClient.List(ctx, &projectList); err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}

	return projectList.Items, nil
}
