package test

import (
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

func (m MockCodebaseService) UpdateDescription(name, description string) error {
	return m.UpdateDescriptionError
}
