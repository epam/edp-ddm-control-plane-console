package test

import (
	"context"

	projectsV1 "github.com/openshift/api/project/v1"
	"github.com/stretchr/testify/mock"
)

type MockProjectsService struct {
	mock.Mock
}

func (p *MockProjectsService) GetAll(ctx context.Context) ([]projectsV1.Project, error) {
	called := p.Called(ctx)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).([]projectsV1.Project), nil
}

func (p *MockProjectsService) Get(ctx context.Context, name string) (*projectsV1.Project, error) {
	called := p.Called(ctx, name)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*projectsV1.Project), nil
}
