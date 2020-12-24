package test

import "ddm-admin-console/models"

type MockRegistryService struct {
	ListResult []*models.Registry
	ListError  error

	CreateResult *models.Registry
	CreateError  error

	GetResult *models.Registry
	GetError  error

	EditDescriptionError error
}

func (m MockRegistryService) List() ([]*models.Registry, error) {
	return m.ListResult, m.ListError
}

func (m MockRegistryService) Create(name, description string) (*models.Registry, error) {
	return m.CreateResult, m.CreateError
}

func (m MockRegistryService) Get(name string) (*models.Registry, error) {
	return m.GetResult, m.GetError
}

func (m MockRegistryService) EditDescription(name, description string) error {
	return m.EditDescriptionError
}
