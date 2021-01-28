package edpcomponent

import (
	"ddm-admin-console/util/consts"

	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type ServiceK8S struct {
	k8sClient rest.Interface
}

func MakeServiceK8S(k8sClient rest.Interface) *ServiceK8S {
	return &ServiceK8S{
		k8sClient: k8sClient,
	}
}

func (s *ServiceK8S) GetAll(namespace string) ([]v1alpha1.EDPComponent, error) {
	var edpComponentsList v1alpha1.EDPComponentList
	if err := s.k8sClient.Get().Namespace(namespace).Resource(consts.EDPComponentPlural).Do().
		Into(&edpComponentsList); err != nil {
		return nil, errors.Wrap(err, "unable to get edp components list")
	}

	return edpComponentsList.Items, nil
}

func (s *ServiceK8S) Get(namespace, name string) (*v1alpha1.EDPComponent, error) {
	var edpComponent v1alpha1.EDPComponent
	if err := s.k8sClient.Get().Namespace(namespace).Resource(consts.EDPComponentPlural).Name(name).Do().
		Into(&edpComponent); err != nil {
		return nil, errors.Wrap(err, "unable to get edp component")
	}

	return &edpComponent, nil
}
