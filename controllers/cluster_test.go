package controllers

import (
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/test"
	"ddm-admin-console/util"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/session"
	"github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	v1alpha12 "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func initBeegoCtrl() (*httptest.ResponseRecorder, beego.Controller) {
	rw := httptest.NewRecorder()
	return rw, beego.Controller{
		Data: map[interface{}]interface{}{},
		Ctx: &context.Context{
			Input: &context.BeegoInput{
				CruSession: &session.MemSessionStore{},
			},
			ResponseWriter: &context.Response{
				ResponseWriter: rw,
			},
			Request: httptest.NewRequest("", "/", nil),
		},
	}
}

func TestClusterManagement_CreateCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("CreateCodebase", command.CreateCodebase{
		Name: "cluster-management", DefaultBranch: "master", Strategy: "import", Lang: "other", BuildTool: "gitops",
		Type: "autotests", Repository: &command.Repository{URL: beego.AppConfig.String("clusterManagementRepo")},
		Description:     util.GetStringP(codebaseDescription),
		Versioning:      command.Versioning{StartFrom: util.GetStringP("0.0.1"), Type: "edp"},
		JobProvisioning: util.GetStringP("default"), GitServer: "gerrit", GitURLPath: util.GetStringP("/cluster-mgmt"),
		JenkinsSlave: util.GetStringP("gitops"), DeploymentScript: "openshift-template", CiTool: "Jenkins",
	}).Return(&v1alpha1.Codebase{ObjectMeta: v1.ObjectMeta{Name: "cb1"}}, nil)
	codebaseService.On("GetCodebaseByNameK8s",
		"cluster-management").Return(nil, service.RegistryNotFound{}).Once()
	codebaseService.On("GetCodebaseByNameK8s",
		"cluster-management").Return(&query.Codebase{Name: "clus1", CodebaseBranch: []*query.CodebaseBranch{
		{
			Name: "br1",
		},
	}}, nil).Once()

	ecs := test.MockEDPComponentServiceK8S{}
	ecs.On("Get", "mdtuddm", "jenkins").Return(&v1alpha12.EDPComponent{}, nil)
	ecs.On("Get", "mdtuddm", "gerrit").Return(&v1alpha12.EDPComponent{}, nil)

	beego.Router("/cluster-management-create", MakeClusterManagement(&codebaseService, &ecs,
		"cluster-management", beego.AppConfig.String("clusterManagementRepo")))
	request, _ := http.NewRequest("GET", "/cluster-management-create", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("cluster management return wrong response code")
	}
}

func TestClusterManagement_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebaseByNameK8s", "cluster-management").
		Return(&query.Codebase{CodebaseBranch: []*query.CodebaseBranch{{}}}, nil)

	ecs := test.MockEDPComponentServiceK8S{}
	ecs.On("Get", "mdtuddm", "jenkins").Return(&v1alpha12.EDPComponent{}, nil)
	ecs.On("Get", "mdtuddm", "gerrit").Return(&v1alpha12.EDPComponent{}, nil)

	beego.Router("/cluster-management", MakeClusterManagement(&codebaseService, &ecs,
		"cluster-management", beego.AppConfig.String("clusterManagementRepo")))
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

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebaseByNameK8s", "cluster-management").
		Return(nil, mockErr)
	ecs := test.MockEDPComponentServiceK8S{}

	beego.Router("/cluster-management-error1", MakeClusterManagement(&codebaseService, &ecs, "cluster-management", ""))
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

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebaseByNameK8s", "cluster-management").Return(&query.Codebase{
		CodebaseBranch: []*query.CodebaseBranch{
			{},
		}}, nil)
	ecs := test.MockEDPComponentServiceK8S{}
	ecs.On("Get", "mdtuddm", "jenkins").Return(nil, mockErr)

	beego.Router("/cluster-management-error2", MakeClusterManagement(&codebaseService, &ecs, "cluster-management", ""))
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
