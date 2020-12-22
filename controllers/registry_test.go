package controllers

import (
	"bytes"
	"ddm-admin-console/service"
	_ "ddm-admin-console/templatefunction"
	"ddm-admin-console/test"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/astaxie/beego"
	v1 "k8s.io/api/core/v1"
)

func TestListRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	k8sMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			ListResult: &v1.NamespaceList{},
		},
	}

	beego.Router("/list-registry", MakeListRegistry(service.MakeRegistry(k8sMock, "test-env")))
	request, _ := http.NewRequest("GET", "/list-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Fatal("list registry return wrong response code")
	}
}

func TestListRegistry_GetFailure(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	k8sMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			//ListResult: &v1.NamespaceList{},
			ListError: errors.New("list error"),
		},
	}

	beego.Router("/list-registry-failure", MakeListRegistry(service.MakeRegistry(k8sMock, "test-env")))
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

	k8sMock := test.MockCoreClient{}
	beego.Router("/create-registry", MakeCreateRegistry(service.MakeRegistry(k8sMock, "test-env")))
	request, _ := http.NewRequest("GET", "/create-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Fatal("create registry not found")
	}
}

func TestCreatRegistry_Post(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	k8sMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{},
		MockConfigMapInterface: test.MockConfigMapInterface{},
	}
	ctrl := MakeCreateRegistry(service.MakeRegistry(k8sMock, "test-env"))
	beego.Router("/create-registry-failure", ctrl)
	request, _ := http.NewRequest("POST", "/create-registry-failure", bytes.NewReader([]byte{}))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("create registry not found")
	}
}

func TestEditRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	beego.ErrorController(&ErrorController{})
	beego.Router("/edit-registry", &EditRegistry{})
	request, _ := http.NewRequest("GET", "/edit-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Fatal("edit registry not found")
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
