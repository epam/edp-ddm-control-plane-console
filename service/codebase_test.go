package service

import (
	"ddm-admin-console/k8s"
	"ddm-admin-console/models"
	"ddm-admin-console/models/query"
	"ddm-admin-console/repository/mock"
	"ddm-admin-console/test"
	"fmt"
	"net/http"
	"testing"

	"k8s.io/client-go/rest"

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

	reg := models.Registry{Name: "foo", Description: "bar"}

	if err := svc.UpdateDescription(&reg); err != nil {
		t.Fatal(fmt.Sprintf("%+v", err))
	}
}

func TestCodebaseService_GetCodebasesByCriteriaK8s(t *testing.T) {
	getHTTPClient := test.MockHTTPClient{
		DoHandler: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/codebases" {
				getResponse, _ := test.NewResponseShortcut(
					fmt.Sprintf(
						`{"items": [{"metadata": {"name": "test"}, "spec": {"description": "test", "type": "%s"}}]}`,
						query.Registry))
				return getResponse, nil
			}

			getResponse, _ := test.NewResponseShortcut(
				`{"items": [{"metadata": {"name": "test"}, "spec": {"branchName": "master"}}]}`)
			return getResponse, nil
		},
	}

	cs := CodebaseService{Clients: k8s.ClientSet{
		EDPRestClient: test.MockRestInterface{
			GetResponseHandler: func() *rest.Request {
				return test.NewRequestShortcut(getHTTPClient)
			},
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

	reg := models.Registry{Name: "foo", Description: "bar"}
	err := svc.UpdateDescription(&reg)
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

	reg := models.Registry{Name: "foo", Description: "bar"}

	err := svc.UpdateDescription(&reg)
	if err == nil {
		t.Fatal("no error on k8s fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Log(fmt.Sprintf("%+v", err))
		t.Fatal("wrong error returned")
	}
}
