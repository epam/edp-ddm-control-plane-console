package controllers

import (
	"bytes"
	"ddm-admin-console/models"
	"ddm-admin-console/service"
	_ "ddm-admin-console/templatefunction"
	"ddm-admin-console/test"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/astaxie/beego"
)

func TestListRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{}

	beego.Router("/list-registry", MakeListRegistry(registryServiceMock))
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

	registryServiceMock := test.MockRegistryService{
		ListError: errors.New("error on namespace list"),
	}

	beego.Router("/list-registry-failure", MakeListRegistry(registryServiceMock))
	request, _ := http.NewRequest("GET", "/list-registry-failure", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("no error on list registry fatal")
	}
}

func TestCreatRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{}
	beego.Router("/create-registry-get", MakeCreateRegistry(registryServiceMock))
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

	registryServiceMock := test.MockRegistryService{}
	ctrl := MakeCreateRegistry(registryServiceMock)
	beego.Router("/create-registry-failure", ctrl)
	request, _ := http.NewRequest("POST", "/create-registry-failure", bytes.NewReader([]byte{}))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on validation error")
	}
}

func TestCreatRegistry_Post_NamespaceExists(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{
		CreateError: service.RegistryExistsError{
			Err: errors.New("registry exists"),
		},
	}
	ctrl := MakeCreateRegistry(registryServiceMock)
	beego.Router("/create-registry-k8s-error", ctrl)

	formData := url.Values{
		"name":        []string{"tests"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-k8s-error", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on k8s namespace exists")
	}
}

func TestCreatRegistry_Post_k8sFatal(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{
		CreateError: errors.New("k8s fatal error"),
	}

	ctrl := MakeCreateRegistry(registryServiceMock)
	beego.Router("/create-registry-k8s-fatal", ctrl)

	formData := url.Values{
		"name":        []string{"tests"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-k8s-fatal", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on k8s fatal error")
	}
}

func TestCreatRegistry_Post_ValidationErrorName(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{}
	ctrl := MakeCreateRegistry(registryServiceMock)
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
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{}
	ctrl := MakeCreateRegistry(registryServiceMock)
	beego.Router("/create-registry-success", ctrl)

	formData := url.Values{
		"name":        []string{"test"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-success", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
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

	registryServiceMock := test.MockRegistryService{
		GetError: errors.New("k8s fatal error"),
	}
	ctrl := MakeEditRegistry(registryServiceMock)

	beego.Router("/edit-registry-get-failure/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-get-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostFailure_k8sFatal(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{
		EditDescriptionError: errors.New("k8s fatal"),
	}
	ctrl := MakeEditRegistry(registryServiceMock)

	beego.Router("/edit-registry-failure/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostFailure_LongDescription(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	registryServiceMock := test.MockRegistryService{}
	ctrl := MakeEditRegistry(registryServiceMock)

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

	registryService := test.MockRegistryService{}
	ctrl := MakeEditRegistry(registryService)

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

	registryServiceMock := test.MockRegistryService{
		GetResult: &models.Registry{},
	}
	ctrl := MakeEditRegistry(registryServiceMock)

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

func TestViewRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	beego.ErrorController(&ErrorController{})
	beego.Router("/view-registry", &ViewRegistry{})
	request, _ := http.NewRequest("GET", "/view-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Fatal("view registry not found")
	}
}
