// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"
	gerrit "ddm-admin-console/service/gerrit"

	go_gerrit "github.com/andygrunwald/go-gerrit"

	mock "github.com/stretchr/testify/mock"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// ApproveAndSubmitChange provides a mock function with given fields: changeID, username, email
func (_m *ServiceInterface) ApproveAndSubmitChange(changeID string, username string, email string) error {
	ret := _m.Called(changeID, username, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(changeID, username, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateMergeRequest provides a mock function with given fields: ctx, mr
func (_m *ServiceInterface) CreateMergeRequest(ctx context.Context, mr *gerrit.MergeRequest) error {
	ret := _m.Called(ctx, mr)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *gerrit.MergeRequest) error); ok {
		r0 = rf(ctx, mr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateMergeRequestWithContents provides a mock function with given fields: ctx, mr, contents
func (_m *ServiceInterface) CreateMergeRequestWithContents(ctx context.Context, mr *gerrit.MergeRequest, contents map[string]string) error {
	ret := _m.Called(ctx, mr, contents)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *gerrit.MergeRequest, map[string]string) error); ok {
		r0 = rf(ctx, mr, contents)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateProject provides a mock function with given fields: ctx, name
func (_m *ServiceInterface) CreateProject(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBranchContent provides a mock function with given fields: projectName, branch, fileLocation
func (_m *ServiceInterface) GetBranchContent(projectName string, branch string, fileLocation string) (string, error) {
	ret := _m.Called(projectName, branch, fileLocation)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string) string); ok {
		r0 = rf(projectName, branch, fileLocation)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(projectName, branch, fileLocation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetChangeDetails provides a mock function with given fields: changeID
func (_m *ServiceInterface) GetChangeDetails(changeID string) (*go_gerrit.ChangeInfo, error) {
	ret := _m.Called(changeID)

	var r0 *go_gerrit.ChangeInfo
	if rf, ok := ret.Get(0).(func(string) *go_gerrit.ChangeInfo); ok {
		r0 = rf(changeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*go_gerrit.ChangeInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(changeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFileContents provides a mock function with given fields: ctx, projectName, branch, filePath
func (_m *ServiceInterface) GetFileContents(ctx context.Context, projectName string, branch string, filePath string) (string, error) {
	ret := _m.Called(ctx, projectName, branch, filePath)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) string); ok {
		r0 = rf(ctx, projectName, branch, filePath)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, projectName, branch, filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMergeListCommits provides a mock function with given fields: ctx, changeID, revision
func (_m *ServiceInterface) GetMergeListCommits(ctx context.Context, changeID string, revision string) ([]gerrit.Commit, error) {
	ret := _m.Called(ctx, changeID, revision)

	var r0 []gerrit.Commit
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []gerrit.Commit); ok {
		r0 = rf(ctx, changeID, revision)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gerrit.Commit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, changeID, revision)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMergeRequest provides a mock function with given fields: ctx, name
func (_m *ServiceInterface) GetMergeRequest(ctx context.Context, name string) (*gerrit.GerritMergeRequest, error) {
	ret := _m.Called(ctx, name)

	var r0 *gerrit.GerritMergeRequest
	if rf, ok := ret.Get(0).(func(context.Context, string) *gerrit.GerritMergeRequest); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gerrit.GerritMergeRequest)
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

// GetMergeRequestByChangeID provides a mock function with given fields: ctx, changeID
func (_m *ServiceInterface) GetMergeRequestByChangeID(ctx context.Context, changeID string) (*gerrit.GerritMergeRequest, error) {
	ret := _m.Called(ctx, changeID)

	var r0 *gerrit.GerritMergeRequest
	if rf, ok := ret.Get(0).(func(context.Context, string) *gerrit.GerritMergeRequest); ok {
		r0 = rf(ctx, changeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gerrit.GerritMergeRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, changeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMergeRequestByProject provides a mock function with given fields: ctx, projectName
func (_m *ServiceInterface) GetMergeRequestByProject(ctx context.Context, projectName string) ([]gerrit.GerritMergeRequest, error) {
	ret := _m.Called(ctx, projectName)

	var r0 []gerrit.GerritMergeRequest
	if rf, ok := ret.Get(0).(func(context.Context, string) []gerrit.GerritMergeRequest); ok {
		r0 = rf(ctx, projectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gerrit.GerritMergeRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, projectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMergeRequests provides a mock function with given fields: ctx
func (_m *ServiceInterface) GetMergeRequests(ctx context.Context) ([]gerrit.GerritMergeRequest, error) {
	ret := _m.Called(ctx)

	var r0 []gerrit.GerritMergeRequest
	if rf, ok := ret.Get(0).(func(context.Context) []gerrit.GerritMergeRequest); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gerrit.GerritMergeRequest)
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

// GetProject provides a mock function with given fields: ctx, name
func (_m *ServiceInterface) GetProject(ctx context.Context, name string) (*gerrit.GerritProject, error) {
	ret := _m.Called(ctx, name)

	var r0 *gerrit.GerritProject
	if rf, ok := ret.Get(0).(func(context.Context, string) *gerrit.GerritProject); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gerrit.GerritProject)
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

// GetProjectInfo provides a mock function with given fields: projectName
func (_m *ServiceInterface) GetProjectInfo(projectName string) (*go_gerrit.ProjectInfo, error) {
	ret := _m.Called(projectName)

	var r0 *go_gerrit.ProjectInfo
	if rf, ok := ret.Get(0).(func(string) *go_gerrit.ProjectInfo); ok {
		r0 = rf(projectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*go_gerrit.ProjectInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(projectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProjects provides a mock function with given fields: ctx
func (_m *ServiceInterface) GetProjects(ctx context.Context) ([]gerrit.GerritProject, error) {
	ret := _m.Called(ctx)

	var r0 []gerrit.GerritProject
	if rf, ok := ret.Get(0).(func(context.Context) []gerrit.GerritProject); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gerrit.GerritProject)
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

// GoGerritClient provides a mock function with given fields:
func (_m *ServiceInterface) GoGerritClient() *go_gerrit.Client {
	ret := _m.Called()

	var r0 *go_gerrit.Client
	if rf, ok := ret.Get(0).(func() *go_gerrit.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*go_gerrit.Client)
		}
	}

	return r0
}

// UpdateMergeRequestStatus provides a mock function with given fields: ctx, mr
func (_m *ServiceInterface) UpdateMergeRequestStatus(ctx context.Context, mr *gerrit.GerritMergeRequest) error {
	ret := _m.Called(ctx, mr)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *gerrit.GerritMergeRequest) error); ok {
		r0 = rf(ctx, mr)
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
