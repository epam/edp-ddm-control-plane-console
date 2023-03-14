// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"
	codebase "ddm-admin-console/service/codebase"

	k8s "ddm-admin-console/service/k8s"

	mock "github.com/stretchr/testify/mock"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// CheckIsAllowedToCreate provides a mock function with given fields: k8sService
func (_m *ServiceInterface) CheckIsAllowedToCreate(k8sService k8s.ServiceInterface) (bool, error) {
	ret := _m.Called(k8sService)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(k8s.ServiceInterface) (bool, error)); ok {
		return rf(k8sService)
	}
	if rf, ok := ret.Get(0).(func(k8s.ServiceInterface) bool); ok {
		r0 = rf(k8sService)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(k8s.ServiceInterface) error); ok {
		r1 = rf(k8sService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CheckIsAllowedToUpdate provides a mock function with given fields: codebaseName, k8sService
func (_m *ServiceInterface) CheckIsAllowedToUpdate(codebaseName string, k8sService k8s.ServiceInterface) (bool, error) {
	ret := _m.Called(codebaseName, k8sService)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string, k8s.ServiceInterface) (bool, error)); ok {
		return rf(codebaseName, k8sService)
	}
	if rf, ok := ret.Get(0).(func(string, k8s.ServiceInterface) bool); ok {
		r0 = rf(codebaseName, k8sService)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string, k8s.ServiceInterface) error); ok {
		r1 = rf(codebaseName, k8sService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CheckPermissions provides a mock function with given fields: initial, k8sService
func (_m *ServiceInterface) CheckPermissions(initial []codebase.Codebase, k8sService k8s.ServiceInterface) ([]codebase.WithPermissions, error) {
	ret := _m.Called(initial, k8sService)

	var r0 []codebase.WithPermissions
	var r1 error
	if rf, ok := ret.Get(0).(func([]codebase.Codebase, k8s.ServiceInterface) ([]codebase.WithPermissions, error)); ok {
		return rf(initial, k8sService)
	}
	if rf, ok := ret.Get(0).(func([]codebase.Codebase, k8s.ServiceInterface) []codebase.WithPermissions); ok {
		r0 = rf(initial, k8sService)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]codebase.WithPermissions)
		}
	}

	if rf, ok := ret.Get(1).(func([]codebase.Codebase, k8s.ServiceInterface) error); ok {
		r1 = rf(initial, k8sService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: cb
func (_m *ServiceInterface) Create(cb *codebase.Codebase) error {
	ret := _m.Called(cb)

	var r0 error
	if rf, ok := ret.Get(0).(func(*codebase.Codebase) error); ok {
		r0 = rf(cb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateBranch provides a mock function with given fields: branch
func (_m *ServiceInterface) CreateBranch(branch *codebase.CodebaseBranch) error {
	ret := _m.Called(branch)

	var r0 error
	if rf, ok := ret.Get(0).(func(*codebase.CodebaseBranch) error); ok {
		r0 = rf(branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateDefaultBranch provides a mock function with given fields: cb
func (_m *ServiceInterface) CreateDefaultBranch(cb *codebase.Codebase) error {
	ret := _m.Called(cb)

	var r0 error
	if rf, ok := ret.Get(0).(func(*codebase.Codebase) error); ok {
		r0 = rf(cb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateTempSecrets provides a mock function with given fields: cb, k8sService, gerritCreatorSecretName
func (_m *ServiceInterface) CreateTempSecrets(cb *codebase.Codebase, k8sService k8s.ServiceInterface, gerritCreatorSecretName string) error {
	ret := _m.Called(cb, k8sService, gerritCreatorSecretName)

	var r0 error
	if rf, ok := ret.Get(0).(func(*codebase.Codebase, k8s.ServiceInterface, string) error); ok {
		r0 = rf(cb, k8sService, gerritCreatorSecretName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: name
func (_m *ServiceInterface) Delete(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: name
func (_m *ServiceInterface) Get(name string) (*codebase.Codebase, error) {
	ret := _m.Called(name)

	var r0 *codebase.Codebase
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*codebase.Codebase, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) *codebase.Codebase); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebase.Codebase)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByType provides a mock function with given fields: tp
func (_m *ServiceInterface) GetAllByType(tp string) ([]codebase.Codebase, error) {
	ret := _m.Called(tp)

	var r0 []codebase.Codebase
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]codebase.Codebase, error)); ok {
		return rf(tp)
	}
	if rf, ok := ret.Get(0).(func(string) []codebase.Codebase); ok {
		r0 = rf(tp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]codebase.Codebase)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBranchesByCodebase provides a mock function with given fields: codebaseName
func (_m *ServiceInterface) GetBranchesByCodebase(codebaseName string) ([]codebase.CodebaseBranch, error) {
	ret := _m.Called(codebaseName)

	var r0 []codebase.CodebaseBranch
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]codebase.CodebaseBranch, error)); ok {
		return rf(codebaseName)
	}
	if rf, ok := ret.Get(0).(func(string) []codebase.CodebaseBranch); ok {
		r0 = rf(codebaseName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]codebase.CodebaseBranch)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(codebaseName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceForContext provides a mock function with given fields: ctx
func (_m *ServiceInterface) ServiceForContext(ctx context.Context) (codebase.ServiceInterface, error) {
	ret := _m.Called(ctx)

	var r0 codebase.ServiceInterface
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (codebase.ServiceInterface, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) codebase.ServiceInterface); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(codebase.ServiceInterface)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: cb
func (_m *ServiceInterface) Update(cb *codebase.Codebase) error {
	ret := _m.Called(cb)

	var r0 error
	if rf, ok := ret.Get(0).(func(*codebase.Codebase) error); ok {
		r0 = rf(cb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
