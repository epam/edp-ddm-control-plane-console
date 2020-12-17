package controllers

import (
	_ "ddm-admin-console/templatefunction"
	"github.com/astaxie/beego"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListRegistry_Get(t *testing.T) {
	if err := os.Chdir(".."); err != nil {
		t.Fatal("unable to change dir")
	}

	beego.TestBeegoInit(".")
	beego.ErrorController(&ErrorController{})
	beego.Router("/list-registry", &ListRegistry{})
	request, _ := http.NewRequest("GET", "/list-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Body.String())
		t.Fatal("list registry not found")
	}
}
