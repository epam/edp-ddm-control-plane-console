package test

import "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"

type MockEDPComponentServiceK8S struct {
	GetAllResult []v1alpha1.EDPComponent
	GetAllError  error

	GetResult *v1alpha1.EDPComponent
	GetError  error
}

func (m MockEDPComponentServiceK8S) GetAll(namespace string) ([]v1alpha1.EDPComponent, error) {
	return m.GetAllResult, m.GetAllError
}

func (m MockEDPComponentServiceK8S) Get(namespace, name string) (*v1alpha1.EDPComponent, error) {
	return m.GetResult, m.GetError
}
