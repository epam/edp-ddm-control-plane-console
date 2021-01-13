package controllers

import (
	"ddm-admin-console/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/astaxie/beego"
)

func TestDashboardController_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	beego.Router("/dashboard-view", MakeDashboardController())

	request, _ := http.NewRequest("GET", "/dashboard-view", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on dashboard controller")
	}
}
