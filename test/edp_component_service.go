package test

import "ddm-admin-console/models/query"

type MockEDPComponentService struct {
	GetEDPComponentResult *query.EDPComponent
	GetEDPComponentError  error
}

func (m MockEDPComponentService) GetEDPComponent(componentType string) (*query.EDPComponent, error) {
	return m.GetEDPComponentResult, m.GetEDPComponentError
}
