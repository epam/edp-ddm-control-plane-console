package controllers

import (
	"bytes"
	"context"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	"ddm-admin-console/test"
	"ddm-admin-console/util"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	"github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	v1alpha12 "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	projectsV1 "github.com/openshift/api/project/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebasesByCriteriaK8s", query.CodebaseCriteria{Type: query.Registry}).
		Return([]*query.Codebase{
			{
				Name:               "name1",
				ForegroundDeletion: false,
			},
			{
				Name:               "name2",
				ForegroundDeletion: false,
			},
		}, nil)

	projectSVC := test.MockProjectsService{}
	projectSVC.On("GetAll", context.Background()).
		Return([]projectsV1.Project{{ObjectMeta: v1.ObjectMeta{Name: "name1"}}}, nil)

	beego.Router("/list-registry", MakeListRegistry("/", "creator", &codebaseService, &projectSVC))
	request, _ := http.NewRequest("GET", "/list-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("list registry return wrong response code")
	}
}

func TestListRegistry_GetFailure(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebasesByCriteriaK8s", query.CodebaseCriteria{Type: "library"}).
		Return(nil, errors.New("error on codebase list"))

	projectSVC := test.MockProjectsService{}

	beego.Router("/list-registry-failure", MakeListRegistry("/", "creator", &codebaseService, &projectSVC))
	request, _ := http.NewRequest("GET", "/list-registry-failure", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("no error on list registry fatal")
	}
}

func TestCreateRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	beego.Router("/create-registry-get", MakeCreateRegistry("/", "creator", &codebaseService))
	request, _ := http.NewRequest("GET", "/create-registry-get", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on registry create get")
	}
}

func TestCreatRegistry_Post_ValidationError(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	grv := mockGroupValidator{}
	grv.On("IsAllowedToCreateRegistry", "creator").Return(true)
	codebaseService := test.MockCodebaseService{}
	ctrl := MakeCreateRegistry("/", "creator", &codebaseService)
	ctrl.GroupValidator = &grv

	beego.Router("/create-registry-failure", ctrl)
	request, _ := http.NewRequest("POST", "/create-registry-failure", bytes.NewReader([]byte{}))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on validation error")
	}
}

func TestCreatRegistry_Post_CodebaseExists(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("CreateCodebase", command.CreateCodebase{
		Name: "name1", DefaultBranch: "master", Strategy: "clone", Lang: "other", BuildTool: "gitops", Type: string(query.Registry),
		Repository:  &command.Repository{URL: beego.AppConfig.String("registryGitRepo")},
		Description: util.GetStringP("desc1"), GitServer: "gerrit",
		Versioning:   command.Versioning{Type: "edp", StartFrom: util.GetStringP("0.0.1")},
		JenkinsSlave: util.GetStringP("gitops"), JobProvisioning: util.GetStringP("default"),
		DeploymentScript: "openshift-template", CiTool: "Jenkins",
	}).
		Return(nil, edperror.NewCodebaseAlreadyExistsError())
	grv := mockGroupValidator{}
	grv.On("IsAllowedToCreateRegistry", "creator").Return(true)

	ctrl := MakeCreateRegistry("/", "creator", &codebaseService)
	ctrl.GroupValidator = &grv
	beego.Router("/create-registry-k8s-error", ctrl)

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "name1"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("description", "desc1"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("sign-key-issuer", "issuer"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("sign-key-pwd", "pwd"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("key6", "fake"); err != nil {
		t.Fatal(err)
	}

	f1, err := writer.CreateFormFile("key6", "key6.dat")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	f1, err = writer.CreateFormFile("ca-cert", "file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	f1, err = writer.CreateFormFile("ca-json", "file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request, _ := http.NewRequest("POST", "/create-registry-k8s-error", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	//request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on k8s namespace exists")
	}
}

func TestCreatRegistry_Post_ValidationErrorName(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	grv := mockGroupValidator{}
	grv.On("IsAllowedToCreateRegistry", "creator").Return(true)
	ctrl := MakeCreateRegistry("/", "creator", &test.MockCodebaseService{})
	ctrl.GroupValidator = &grv
	beego.Router("/create-registry-error-name", ctrl)

	formData := url.Values{
		"name":        []string{"test!s"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-error-name", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on name validation")
	}
}

func TestCreatRegistry_Post_Success(t *testing.T) {
	codebaseService := test.MockCodebaseService{}

	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	grv := mockGroupValidator{}
	grv.On("IsAllowedToCreateRegistry", "creator").Return(true)

	ctrl := MakeCreateRegistry("/", "creator", &codebaseService)
	ctrl.GroupValidator = &grv
	beego.Router("/create-registry-success", ctrl)

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "name1"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("description", "desc1"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("sign-key-issuer", "issuer"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("sign-key-pwd", "pwd"); err != nil {
		t.Fatal(err)
	}

	if err := writer.WriteField("key6", "fake"); err != nil {
		t.Fatal(err)
	}

	f1, err := writer.CreateFormFile("key6", "key6.dat")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	f1, err = writer.CreateFormFile("ca-cert", "file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	f1, err = writer.CreateFormFile("ca-json", "file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f1.Write([]byte("test data")); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	codebaseService.On("CreateCodebase", command.CreateCodebase{
		Name: "name1", DefaultBranch: "master", Strategy: "clone", Lang: "other", BuildTool: "gitops", Type: string(query.Registry),
		Repository:  &command.Repository{URL: beego.AppConfig.String("registryGitRepo")},
		Description: util.GetStringP("desc1"), GitServer: "gerrit",
		Versioning:   command.Versioning{Type: "edp", StartFrom: util.GetStringP("0.0.1")},
		JenkinsSlave: util.GetStringP("gitops"), JobProvisioning: util.GetStringP("default"),
		DeploymentScript: "openshift-template", CiTool: "Jenkins",
	}).
		Return(&v1alpha1.Codebase{}, nil)

	request, _ := http.NewRequest("POST", "/create-registry-success", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 303 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on namespace creation")
	}
}

func TestEditRegistry_GetFailure(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("k8s fatal error")
	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebaseByNameK8s", "test").
		Return(nil, mockErr)

	projectsSvc := test.MockProjectsService{}
	jenkinsSvc := test.MockJenkinsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)

	ctrl := MakeEditRegistry(&codebaseService, &projectsSvc, &jenkinsSvc)

	beego.Router("/edit-registry-get-failure/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-get-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong body response")
	}
}

func TestEditRegistry_GetFailure404(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	codebaseService.On("GetCodebaseByNameK8s", "test").Return(nil,
		service.RegistryNotFound{})

	projectsSvc := test.MockProjectsService{}
	jenkinsSvc := test.MockJenkinsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)

	ctrl := MakeEditRegistry(&codebaseService, &projectsSvc, &jenkinsSvc)

	beego.Router("/edit-registry-get-failure404/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-get-failure404/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}

	if !strings.Contains(responseWriter.Body.String(), "Sorry, page not found") {
		t.Fatal("no error in response body")
	}
}

func TestEditRegistry_PostFailure_k8sFatal(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("k8s fatal")
	cbMock := test.MockCodebaseService{}
	cbMock.On("UpdateDescription", &models.Registry{Name: "test"}).Return(mockErr)
	cbMock.On("GetCodebaseByNameK8s", "test").Return(&query.Codebase{}, nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)
	jenkinsSvc := test.MockJenkinsService{}

	ctrl := MakeEditRegistry(&cbMock, &projectsSvc, &jenkinsSvc)

	beego.Router("/edit-registry-failure/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong body response")
	}
}

func TestEditRegistry_PostFailure_LongDescription(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "test").Return(&query.Codebase{}, nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)
	jenkinsSvc := test.MockJenkinsService{}

	ctrl := MakeEditRegistry(&cbMock, &projectsSvc, &jenkinsSvc)

	formData := url.Values{
		"description": []string{`test11111111111111111111111111111111111111111111111111111111111111111111111test1111111
1111111111111111111111111111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111
111111111111111111111test11111111111111111111111111111111111111111111111111111111111111111111111test1111111111111111111
1111111111111111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111111111111111
111111111test11111111111111111111111111111111111111111111111111111111111111111111111test111111111111111111111111111111
11111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111111111111111111111111t
est11111111111111111111111111111111111111111111111111111111111111111111111test111111111111111111111111111111111111111
11111111111111111111111111111111`},
	}

	beego.Router("/edit-registry-failure-description/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-failure-description/test", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("UpdateDescription", &models.Registry{Name: "test", Description: "test1"}).
		Return(nil)
	cbMock.On("GetCodebaseByNameK8s", "test").
		Return(&query.Codebase{}, nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)
	jenkinsSvc := test.MockJenkinsService{}
	var params map[string]string
	jenkinsSvc.On("CreateJobBuildRun", "test/job/MASTER-Build-test/", params).Return(nil)
	ctrl := MakeEditRegistry(&cbMock, &projectsSvc, &jenkinsSvc)

	formData := url.Values{
		"description": []string{"test1"},
	}

	beego.Router("/edit-registry-success-description/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-success-description/test", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 303 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit success")
	}
}

func TestEditRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "test").Return(&query.Codebase{}, nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "test").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "test"}}, nil)
	jenkinsSvc := test.MockJenkinsService{}

	ctrl := MakeEditRegistry(&cbMock, &projectsSvc, &jenkinsSvc)

	beego.Router("/edit-registry-success/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-success/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit")
	}
}

func TestListRegistry_DeleteRegistry_FailureGetCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("GetCodebaseByNameError fatal")
	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").Return(nil, mockErr)

	projectSVC := test.MockProjectsService{}
	projectSVC.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	listRegistryCtrl := MakeListRegistry("/", "creator", &cbMock, &projectSVC)

	beego.Router("/delete-registry-FailureGetCodebase", listRegistryCtrl)
	v := url.Values{"registry-name": []string{"name1"}}
	buffer := bytes.Buffer{}
	buffer.WriteString(v.Encode())
	request, _ := http.NewRequest("POST", "/delete-registry-FailureGetCodebase", &buffer)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on delete registry")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}

func TestListRegistry_DeleteRegistry_FailureGetCodebase404(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").
		Return(nil, errors.Wrap(service.RegistryNotFound{}, ""))

	projectSVC := test.MockProjectsService{}
	projectSVC.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	listRegistryCtrl := MakeListRegistry("/", "creator", &cbMock, &projectSVC)

	v := url.Values{"registry-name": []string{"name1"}}
	buffer := bytes.Buffer{}
	buffer.WriteString(v.Encode())

	beego.Router("/delete-registry-FailureGetCodebase404", listRegistryCtrl)
	request, _ := http.NewRequest("POST", "/delete-registry-FailureGetCodebase404", &buffer)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on delete registry")
	}

	if !strings.Contains(responseWriter.Body.String(), "Sorry, page not found") {
		t.Log(responseWriter.Body.String())
		t.Fatal("no error in response body")
	}
}

func TestListRegistry_DeleteRegistry_FailureDeleteCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("DeleteCodebase fatal")
	cbMock := test.MockCodebaseService{}

	projectSVC := test.MockProjectsService{}
	projectSVC.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	listRegistryCtrl := MakeListRegistry("/", "creator", &cbMock, &projectSVC)
	cbMock.On("GetCodebaseByNameK8s", "name1").Return(&query.Codebase{Type: "t1", Name: "n1"}, nil)
	cbMock.On("Delete", "n1", "t1").Return(mockErr)

	beego.Router("/delete-registry-DeleteCodebase", listRegistryCtrl)
	v := url.Values{"registry-name": []string{"name1"}}
	buffer := bytes.Buffer{}
	buffer.WriteString(v.Encode())
	request, _ := http.NewRequest("POST", "/delete-registry-DeleteCodebase", &buffer)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on delete registry")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}

func TestListRegistry_DeleteRegistry(t *testing.T) {
	rw, ctrl := initBeegoCtrl()
	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").Return(&query.Codebase{Name: "name2", Type: "type1"}, nil)
	cbMock.On("Delete", "name2", "type1").Return(nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	listRegistryCtrl := MakeListRegistry("/", "creator", &cbMock, &projectsSvc)
	ctrl.Ctx.Input.SetParam("registry-name", "name1")
	listRegistryCtrl.Controller = ctrl

	listRegistryCtrl.Post()

	if rw.Code != 303 {
		t.Log(rw.Code)
		t.Fatal("wrong response code on delete registry")
	}
}

func TestViewRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").
		Return(&query.Codebase{CodebaseBranch: []*query.CodebaseBranch{{}}, ActionLog: []*query.ActionLog{{}}}, nil)

	eds := test.MockEDPComponentServiceK8S{}
	eds.On("Get", "mdtuddm", "jenkins").Return(&v1alpha12.EDPComponent{}, nil)
	eds.On("Get", "mdtuddm", "gerrit").Return(&v1alpha12.EDPComponent{}, nil)
	eds.On("GetAll", "").Return([]v1alpha12.EDPComponent{{}}, nil)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	beego.Router("/view-registry/:name", MakeViewRegistry(&cbMock, &eds, &projectsSvc, "", "mdtuddm"))
	request, _ := http.NewRequest("GET", "/view-registry/name1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}
}

func TestViewRegistry_Get_FailureEdpComponents(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").
		Return(&query.Codebase{CodebaseBranch: []*query.CodebaseBranch{{}}, ActionLog: []*query.ActionLog{{}}}, nil)

	mockErr := errors.New("GetEDPComponents fatal")

	eds := test.MockEDPComponentServiceK8S{}
	eds.On("Get", "mdtuddm", "jenkins").Return(&v1alpha12.EDPComponent{}, nil)
	eds.On("Get", "mdtuddm", "gerrit").Return(&v1alpha12.EDPComponent{}, nil)
	eds.On("GetAll", "").Return(nil, mockErr)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	beego.Router("/view-registry-failure-edp-comp/:name", MakeViewRegistry(&cbMock, &eds, &projectsSvc, "", "mdtuddm"))
	request, _ := http.NewRequest("GET", "/view-registry-failure-edp-comp/name1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("wrong error return in response body")
	}
}

func TestViewRegistry_Get_FailureGetCodebaseByName(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}
	mockErr := errors.New("GetCodebaseByName fatal")

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").Return(nil, mockErr)
	eds := test.MockEDPComponentServiceK8S{}

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	beego.Router("/view-registry-FailureGetCodebaseByName/:name", MakeViewRegistry(&cbMock, &eds, &projectsSvc, "", "mdtuddm"))
	request, _ := http.NewRequest("GET", "/view-registry-FailureGetCodebaseByName/name1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("wrong error return in response body")
	}
}

func TestViewRegistry_Get_FailureGetCodebaseByName404(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").
		Return(nil, errors.Wrap(service.RegistryNotFound{}, ""))
	eds := test.MockEDPComponentServiceK8S{}

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	beego.Router("/view-registry-FailureGetCodebaseByName404/:name",
		MakeViewRegistry(&cbMock, &eds, &projectsSvc, "", "mdtuddm"))
	request, _ := http.NewRequest("GET", "/view-registry-FailureGetCodebaseByName404/name1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), "Sorry, page not found") {
		t.Fatal("wrong error return in response body")
	}
}

func TestViewRegistry_Get_FailureCreateLinksForGerritProvider(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}
	mockErr := errors.New("GetEDPComponentError fatal")

	cbMock := test.MockCodebaseService{}
	cbMock.On("GetCodebaseByNameK8s", "name1").
		Return(&query.Codebase{CodebaseBranch: []*query.CodebaseBranch{{}}}, nil)

	eds := test.MockEDPComponentServiceK8S{}
	eds.On("Get", "mdtuddm", "jenkins").Return(nil, mockErr)

	projectsSvc := test.MockProjectsService{}
	projectsSvc.On("Get", context.Background(), "name1").
		Return(&projectsV1.Project{ObjectMeta: v1.ObjectMeta{Name: "name1"}}, nil)

	beego.Router("/view-registry-FailureCreateLinksForGerritProvider/:name",
		MakeViewRegistry(&cbMock, &eds, &projectsSvc, "", "mdtuddm"))
	request, _ := http.NewRequest("GET", "/view-registry-FailureCreateLinksForGerritProvider/name1", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("wrong error return in response body")
	}
}
