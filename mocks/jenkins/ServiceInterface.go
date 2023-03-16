// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	context "context"
	jenkins "ddm-admin-console/service/jenkins"

	mock "github.com/stretchr/testify/mock"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// CreateJobBuildRun provides a mock function with given fields: ctx, name, jobPath, jobParams
func (_m *ServiceInterface) CreateJobBuildRun(ctx context.Context, name string, jobPath string, jobParams map[string]string) error {
	ret := _m.Called(ctx, name, jobPath, jobParams)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, map[string]string) error); ok {
		r0 = rf(ctx, name, jobPath, jobParams)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateJobBuildRunRaw provides a mock function with given fields: ctx, jb
func (_m *ServiceInterface) CreateJobBuildRunRaw(ctx context.Context, jb *jenkins.JenkinsJobBuildRun) error {
	ret := _m.Called(ctx, jb)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *jenkins.JenkinsJobBuildRun) error); ok {
		r0 = rf(ctx, jb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetJobStatus provides a mock function with given fields: ctx, jobName
func (_m *ServiceInterface) GetJobStatus(ctx context.Context, jobName string) (string, int64, error) {
	ret := _m.Called(ctx, jobName)

	var r0 string
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, int64, error)); ok {
		return rf(ctx, jobName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, jobName)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) int64); ok {
		r1 = rf(ctx, jobName)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, jobName)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ServiceForContext provides a mock function with given fields: ctx
func (_m *ServiceInterface) ServiceForContext(ctx context.Context) (jenkins.ServiceInterface, error) {
	ret := _m.Called(ctx)

	var r0 jenkins.ServiceInterface
	if rf, ok := ret.Get(0).(func(context.Context) jenkins.ServiceInterface); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jenkins.ServiceInterface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewServiceInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewServiceInterface creates a new instance of ServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceInterface(t mockConstructorTestingTNewServiceInterface) *ServiceInterface {
	mock := &ServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
