package test

import (
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"

	"github.com/stretchr/testify/mock"

	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
)

type MockCodebaseService struct {
	mock.Mock
}

func (m *MockCodebaseService) CreateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error {
	return nil
}

func (m *MockCodebaseService) GetCodebaseByNameK8s(name string) (*query.Codebase, error) {
	args := m.Called(name)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).(*query.Codebase), nil
}

func (m *MockCodebaseService) GetCodebasesByCriteriaK8s(criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	args := m.Called(criteria)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).([]*query.Codebase), nil
}

func (m *MockCodebaseService) Delete(name, codebaseType string) error {
	return m.Called(name, codebaseType).Error(0)
}

func (m *MockCodebaseService) CreateCodebase(codebase command.CreateCodebase) (*edpv1alpha1.Codebase, error) {
	args := m.Called(codebase)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).(*edpv1alpha1.Codebase), nil
}

func (m *MockCodebaseService) GetCodebasesByCriteria(criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	args := m.Called(criteria)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).([]*query.Codebase), nil
}

func (m *MockCodebaseService) GetCodebaseByName(name string) (*query.Codebase, error) {
	args := m.Called(name)
	if err := args.Error(1); err != nil {
		return nil, err
	}

	return args.Get(0).(*query.Codebase), nil
}

func (m *MockCodebaseService) UpdateDescription(reg *models.Registry) error {
	return m.Called(reg).Error(0)
}

func (m *MockCodebaseService) ExistCodebaseAndBranch(cbName, brName string) bool {
	return m.Called(cbName, brName).Bool(0)
}

func (m *MockCodebaseService) UpdateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error {
	return m.Called(key6, caCert, casJSON, signKeyIssuer, signKeyPwd, registryName).Error(0)
}
