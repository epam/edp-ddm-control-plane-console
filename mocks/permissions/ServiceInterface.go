// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	codebase "ddm-admin-console/service/codebase"

	gin "github.com/gin-gonic/gin"

	k8s "ddm-admin-console/service/k8s"

	mock "github.com/stretchr/testify/mock"

	permissions "ddm-admin-console/service/permissions"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// DeleteRegistry provides a mock function with given fields: name
func (_m *ServiceInterface) DeleteRegistry(name string) {
	_m.Called(name)
}

// DeleteToken provides a mock function with given fields: tok
func (_m *ServiceInterface) DeleteToken(tok string) {
	_m.Called(tok)
}

// DeleteTokenContext provides a mock function with given fields: ctx
func (_m *ServiceInterface) DeleteTokenContext(ctx *gin.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(*gin.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FilterCodebases provides a mock function with given fields: ginContext, cbs, k8sService
func (_m *ServiceInterface) FilterCodebases(ginContext *gin.Context, cbs []codebase.Codebase, k8sService k8s.ServiceInterface) ([]codebase.WithPermissions, error) {
	ret := _m.Called(ginContext, cbs, k8sService)

	var r0 []codebase.WithPermissions
	if rf, ok := ret.Get(0).(func(*gin.Context, []codebase.Codebase, k8s.ServiceInterface) []codebase.WithPermissions); ok {
		r0 = rf(ginContext, cbs, k8sService)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]codebase.WithPermissions)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*gin.Context, []codebase.Codebase, k8s.ServiceInterface) error); ok {
		r1 = rf(ginContext, cbs, k8sService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPermission provides a mock function with given fields: token, registryName
func (_m *ServiceInterface) GetPermission(token string, registryName string) (*permissions.RegistryPermission, error) {
	ret := _m.Called(token, registryName)

	var r0 *permissions.RegistryPermission
	if rf, ok := ret.Get(0).(func(string, string) *permissions.RegistryPermission); ok {
		r0 = rf(token, registryName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*permissions.RegistryPermission)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(token, registryName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadUserRegistries provides a mock function with given fields: ctx
func (_m *ServiceInterface) LoadUserRegistries(ctx *gin.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(*gin.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPermission provides a mock function with given fields: token, registryName, permission
func (_m *ServiceInterface) SetPermission(token string, registryName string, permission permissions.RegistryPermission) {
	_m.Called(token, registryName, permission)
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
