// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"
	edpcomponent "ddm-admin-console/service/edp_component"

	mock "github.com/stretchr/testify/mock"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, name
func (_m *ServiceInterface) Get(ctx context.Context, name string) (*edpcomponent.EDPComponent, error) {
	ret := _m.Called(ctx, name)

	var r0 *edpcomponent.EDPComponent
	if rf, ok := ret.Get(0).(func(context.Context, string) *edpcomponent.EDPComponent); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*edpcomponent.EDPComponent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: ctx, onlyVisible
func (_m *ServiceInterface) GetAll(ctx context.Context, onlyVisible bool) ([]edpcomponent.EDPComponent, error) {
	ret := _m.Called(ctx, onlyVisible)

	var r0 []edpcomponent.EDPComponent
	if rf, ok := ret.Get(0).(func(context.Context, bool) []edpcomponent.EDPComponent); ok {
		r0 = rf(ctx, onlyVisible)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]edpcomponent.EDPComponent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, bool) error); ok {
		r1 = rf(ctx, onlyVisible)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllCategory provides a mock function with given fields: ctx, ns
func (_m *ServiceInterface) GetAllCategory(ctx context.Context, ns string) (map[string][]edpcomponent.EDPComponentItem, error) {
	ret := _m.Called(ctx, ns)

	var r0 map[string][]edpcomponent.EDPComponentItem
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string][]edpcomponent.EDPComponentItem); ok {
		r0 = rf(ctx, ns)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]edpcomponent.EDPComponentItem)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ns)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllCategoryPlatform provides a mock function with given fields: ctx
func (_m *ServiceInterface) GetAllCategoryPlatform(ctx context.Context) (map[string][]edpcomponent.EDPComponentItem, error) {
	ret := _m.Called(ctx)

	var r0 map[string][]edpcomponent.EDPComponentItem
	if rf, ok := ret.Get(0).(func(context.Context) map[string][]edpcomponent.EDPComponentItem); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]edpcomponent.EDPComponentItem)
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

// GetAllNamespace provides a mock function with given fields: ctx, ns, onlyVisible
func (_m *ServiceInterface) GetAllNamespace(ctx context.Context, ns string, onlyVisible bool) ([]edpcomponent.EDPComponent, error) {
	ret := _m.Called(ctx, ns, onlyVisible)

	var r0 []edpcomponent.EDPComponent
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) []edpcomponent.EDPComponent); ok {
		r0 = rf(ctx, ns, onlyVisible)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]edpcomponent.EDPComponent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bool) error); ok {
		r1 = rf(ctx, ns, onlyVisible)
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
