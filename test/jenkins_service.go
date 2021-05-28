package test

import "github.com/stretchr/testify/mock"

type MockJenkinsService struct {
	mock.Mock
}

func (j *MockJenkinsService) CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error {
	return j.Called(jobPath, jobParams).Error(0)
}
