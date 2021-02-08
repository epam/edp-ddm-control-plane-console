package controllers

import (
	"ddm-admin-console/models/query"
	"ddm-admin-console/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/pkg/errors"
)

func TestClusterManagement_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{
		GetCodebaseByNameK8sResult: &query.Codebase{
			CodebaseBranch: []*query.CodebaseBranch{
				{},
			},
		},
	}
	ecs := test.MockEDPComponentServiceK8S{
		GetResult: &v1alpha1.EDPComponent{},
	}

	beego.Router("/cluster-management", MakeClusterManagement(codebaseService, ecs, "cluster-management"))
	request, _ := http.NewRequest("GET", "/cluster-management", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("cluster management return wrong response code")
	}
}

func TestClusterManagement_FailureCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("GetCodebaseByNameError fatal")

	codebaseService := test.MockCodebaseService{
		GetCodebaseByNameK8sError: mockErr,
	}
	ecs := test.MockEDPComponentServiceK8S{
		GetResult: &v1alpha1.EDPComponent{},
	}

	beego.Router("/cluster-management-error1", MakeClusterManagement(codebaseService, ecs, "cluster-management"))
	request, _ := http.NewRequest("GET", "/cluster-management-error1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("cluster management return wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}

func TestClusterManagement_FailureEdpComponent(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("CreateLinksForGerritProviderK8s fatal")

	codebaseService := test.MockCodebaseService{
		GetCodebaseByNameK8sResult: &query.Codebase{
			CodebaseBranch: []*query.CodebaseBranch{
				{},
			},
		},
	}
	ecs := test.MockEDPComponentServiceK8S{
		GetError: mockErr,
	}

	beego.Router("/cluster-management-error2", MakeClusterManagement(codebaseService, ecs, "cluster-management"))
	request, _ := http.NewRequest("GET", "/cluster-management-error2", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("cluster management return wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}
