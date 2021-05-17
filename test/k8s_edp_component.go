package test

import (
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/stretchr/testify/mock"
)

type MockEDPComponentServiceK8S struct {
	mock.Mock
}

func (m *MockEDPComponentServiceK8S) GetAll(namespace string) ([]v1alpha1.EDPComponent, error) {
	args := m.Called(namespace)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).([]v1alpha1.EDPComponent), nil
}

func (m *MockEDPComponentServiceK8S) Get(namespace, name string) (*v1alpha1.EDPComponent, error) {
	args := m.Called(namespace, name)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).(*v1alpha1.EDPComponent), nil
}
