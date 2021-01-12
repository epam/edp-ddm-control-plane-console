package service

import (
	"ddm-admin-console/k8s"
	"ddm-admin-console/models/query"
	"ddm-admin-console/repository/mock"
	"ddm-admin-console/test"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetCodebaseByNameMethod_ShouldBeExecutedSuccessfully(t *testing.T) {
	mCodebase := new(mock.MockCodebase)
	cs := CodebaseService{
		ICodebaseRepository: mCodebase,
	}

	mCodebase.On("GetCodebaseByName", "stub-name").Return(
		query.Codebase{}, nil)

	c, err := cs.GetCodebaseByName("stub-name")
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestGetCodebaseByNameMethod_ShouldBeExecutedWithError(t *testing.T) {
	mCodebase := new(mock.MockCodebase)
	cs := CodebaseService{
		ICodebaseRepository: mCodebase,
	}

	mCodebase.On("GetCodebaseByName", "stub-name").Return(
		nil, errors.New("stub-msg"))

	c, err := cs.GetCodebaseByName("stub-name")
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestCodebaseService_UpdateDescription(t *testing.T) {
	getResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"description": "test"}}`)
	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
	}

	putResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"description": "test"}}`)
	putHTTPClient := test.MockHTTPClient{
		DoResponse: putResponse,
	}

	svc := CodebaseService{
		Clients: k8s.ClientSet{
			EDPRestClient: test.MockRestInterface{
				GetResponse: test.NewRequestShortcut(getHTTPClient),
				PutResponse: test.NewRequestShortcut(putHTTPClient),
			},
		},
	}

	if err := svc.UpdateDescription("foo", "bar"); err != nil {
		t.Fatal(fmt.Sprintf("%+v", err))
	}
}

func TestCodebaseService_GetCodebasesByCriteriaK8s(t *testing.T) {
	// clientSet := k8s.CreateOpenShiftClients()

	getResponse, _ := test.NewResponseShortcut(
		fmt.Sprintf(
			`{"items": [{"metadata": {"name": "test"}, "spec": {"description": "test", "type": "%s"}}]}`,
			query.Registry))

	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    nil,
	}

	cs := CodebaseService{Clients: k8s.ClientSet{
		EDPRestClient: test.MockRestInterface{
			GetResponse: test.NewRequestShortcut(getHTTPClient),
		},
	}}
	rsp, err := cs.GetCodebasesByCriteriaK8s(query.CodebaseCriteria{
		Type: query.Registry,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(rsp) == 0 {
		t.Fatal("no codebases returned")
	}
}

func TestCodebaseService_UpdateDescription_FailureGetCodebase(t *testing.T) {
	mockErr := errors.New("k8s fatal")
	getResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"description": "test"}}`)
	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    mockErr,
	}

	svc := CodebaseService{
		Clients: k8s.ClientSet{
			EDPRestClient: test.MockRestInterface{
				GetResponse: test.NewRequestShortcut(getHTTPClient),
			},
		},
	}

	err := svc.UpdateDescription("foo", "bar")
	if err == nil {
		t.Fatal("no error on k8s fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Log(fmt.Sprintf("%+v", err))
		t.Fatal("wrong error returned")
	}
}

func TestCodebaseService_UpdateDescription_FailureUpdateCodebase(t *testing.T) {
	mockErr := errors.New("k8s fatal")

	getResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"description": "test"}}`)
	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
	}

	putResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"description": "test"}}`)
	putHTTPClient := test.MockHTTPClient{
		DoResponse: putResponse,
		DoError:    mockErr,
	}

	svc := CodebaseService{
		Clients: k8s.ClientSet{
			EDPRestClient: test.MockRestInterface{
				GetResponse: test.NewRequestShortcut(getHTTPClient),
				PutResponse: test.NewRequestShortcut(putHTTPClient),
			},
		},
	}

	err := svc.UpdateDescription("foo", "bar")
	if err == nil {
		t.Fatal("no error on k8s fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Log(fmt.Sprintf("%+v", err))
		t.Fatal("wrong error returned")
	}
}
