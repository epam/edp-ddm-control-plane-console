// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	api "github.com/hashicorp/vault/api"
	mock "github.com/stretchr/testify/mock"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// Read provides a mock function with given fields: path
func (_m *ServiceInterface) Read(path string) (map[string]interface{}, error) {
	ret := _m.Called(path)

	var r0 map[string]interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (map[string]interface{}, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) map[string]interface{}); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadRaw provides a mock function with given fields: path
func (_m *ServiceInterface) ReadRaw(path string) (*api.Secret, error) {
	ret := _m.Called(path)

	var r0 *api.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*api.Secret, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) *api.Secret); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Write provides a mock function with given fields: path, data
func (_m *ServiceInterface) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	ret := _m.Called(path, data)

	var r0 *api.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) (*api.Secret, error)); ok {
		return rf(path, data)
	}
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) *api.Secret); ok {
		r0 = rf(path, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(string, map[string]interface{}) error); ok {
		r1 = rf(path, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteRaw provides a mock function with given fields: path, data
func (_m *ServiceInterface) WriteRaw(path string, data map[string]interface{}) (*api.Secret, error) {
	ret := _m.Called(path, data)

	var r0 *api.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) (*api.Secret, error)); ok {
		return rf(path, data)
	}
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) *api.Secret); ok {
		r0 = rf(path, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(string, map[string]interface{}) error); ok {
		r1 = rf(path, data)
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
