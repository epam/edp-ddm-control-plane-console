package test

import (
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"

	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
)

type MockCodebaseService struct {
	CreateResult *edpv1alpha1.Codebase
	CreateError  error

	GetByCriteriaResult []*query.Codebase
	GetByCriteriaError  error

	UpdateDescriptionError error

	GetCodebaseByNameResult *query.Codebase
	GetCodebaseByNameError  error

	ExistCodebaseAndBranchResult bool

	DeleteError error

	GetCodebasesByCriteriaK8sResult []*query.Codebase
	GetCodebasesByCriteriaK8sError  error

	GetCodebaseByNameK8sResult *query.Codebase
	GetCodebaseByNameK8sError  error

	GetCodebaseByNameK8sMockFunc func(name string) (*query.Codebase, error)
}

func (m MockCodebaseService) CreateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd, registryName string) error {
	return nil
}

func (m MockCodebaseService) GetCodebaseByNameK8s(name string) (*query.Codebase, error) {
	if m.GetCodebaseByNameK8sMockFunc != nil {
		return m.GetCodebaseByNameK8sMockFunc(name)
	}
	return m.GetCodebaseByNameK8sResult, m.GetCodebaseByNameK8sError
}

func (m MockCodebaseService) GetCodebasesByCriteriaK8s(criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	return m.GetCodebasesByCriteriaK8sResult, m.GetCodebasesByCriteriaK8sError
}

func (m MockCodebaseService) Delete(name, codebaseType string) error {
	return m.DeleteError
}

func (m MockCodebaseService) CreateCodebase(codebase command.CreateCodebase) (*edpv1alpha1.Codebase, error) {
	return m.CreateResult, m.CreateError
}

func (m MockCodebaseService) GetCodebasesByCriteria(criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	return m.GetByCriteriaResult, m.GetByCriteriaError
}

func (m MockCodebaseService) GetCodebaseByName(name string) (*query.Codebase, error) {
	return m.GetCodebaseByNameResult, m.GetCodebaseByNameError
}

func (m MockCodebaseService) UpdateDescription(reg *models.Registry) error {
	return m.UpdateDescriptionError
}

func (m MockCodebaseService) ExistCodebaseAndBranch(cbName, brName string) bool {
	return m.ExistCodebaseAndBranchResult
}
