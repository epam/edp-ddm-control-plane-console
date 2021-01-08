package controllers

import (
	"ddm-admin-console/models/query"
	"ddm-admin-console/test"
	"net/http/httptest"
	"testing"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/session"
	"github.com/pkg/errors"
)

func TestDiagramController_getCodebasesJSON(t *testing.T) {
	cs := test.MockCodebaseService{
		GetByCriteriaResult: []*query.Codebase{
			{},
		},
	}
	dc := DiagramController{
		CodebaseService: cs,
	}

	if _, err := dc.getCodebasesJSON(); err != nil {
		t.Fatal(err)
	}
}

func TestDiagramController_getCodebasesJSON_Failure(t *testing.T) {
	mockErr := errors.New("mock err")
	cs := test.MockCodebaseService{
		GetByCriteriaError: mockErr,
	}
	dc := DiagramController{
		CodebaseService: cs,
	}

	_, err := dc.getCodebasesJSON()
	if err == nil {
		t.Fatal("no error on codebase fetch fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Fatal("wrong error type returned")
	}
}

func TestDiagramController_getCodebaseDockerStreamsJSON(t *testing.T) {
	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllCodebaseDockerStreamsResult: []string{"foo", "bar"},
		},
	}

	if _, err := dc.getCodebaseDockerStreamsJSON(); err != nil {
		t.Fatal(err)
	}
}

func TestDiagramController_getCodebaseDockerStreamsJSON_Failure(t *testing.T) {
	mockErr := errors.New("mock err")

	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllCodebaseDockerStreamsError: mockErr,
		},
	}

	_, err := dc.getCodebaseDockerStreamsJSON()

	if err == nil {
		t.Fatal("no error on codebase fetch fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Fatal("wrong error type returned")
	}
}

func TestDiagramController_getPipelinesJSON(t *testing.T) {
	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllPipelinesResult: []*query.CDPipeline{
				{},
			},
		},
	}

	if _, err := dc.getPipelinesJSON(); err != nil {
		t.Fatal(err)
	}
}

func TestDiagramController_getPipelinesJSON_Failure(t *testing.T) {
	mockErr := errors.New("mock err")

	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllPipelinesError: mockErr,
		},
	}

	_, err := dc.getPipelinesJSON()

	if err == nil {
		t.Fatal("no error on codebase fetch fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Fatal("wrong error type returned")
	}
}

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

func TestDiagramController_GetDiagramPage_FailurePipelinesJSON(t *testing.T) {
	rw, ctrl := initBeegoCtrl()

	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllPipelinesError: errors.New("fatal"),
		},
		CodebaseService: test.MockCodebaseService{},
		Controller:      ctrl,
	}

	defer func() {
		r := recover()
		switch r := r.(type) {
		case error:
			if r != beego.ErrAbort {
				t.Fatal(r)
			}
		default:
			t.Fatal(r)
		}

		if rw.Code != 500 {
			t.Fatal("wrong response code")
		}
	}()

	dc.GetDiagramPage()
}

func TestDiagramController_GetDiagramPage_FailureCodebasesJSON(t *testing.T) {
	rw, ctrl := initBeegoCtrl()

	dc := DiagramController{
		PipelineService: test.MockPipelineService{},
		CodebaseService: test.MockCodebaseService{
			GetByCriteriaError: errors.New("fatal"),
		},
		Controller: ctrl,
	}

	defer func() {
		r := recover()
		switch r := r.(type) {
		case error:
			if r != beego.ErrAbort {
				t.Fatal(r)
			}
		default:
			t.Fatal(r)
		}

		if rw.Code != 500 {
			t.Fatal("wrong response code")
		}
	}()

	dc.GetDiagramPage()
}

func TestDiagramController_GetDiagramPage_FailureCodebaseDockerStreamsJSON(t *testing.T) {
	rw, ctrl := initBeegoCtrl()

	dc := DiagramController{
		PipelineService: test.MockPipelineService{
			GetAllCodebaseDockerStreamsError: errors.New("fatal"),
		},
		CodebaseService: test.MockCodebaseService{},
		Controller:      ctrl,
	}

	defer func() {
		r := recover()
		switch r := r.(type) {
		case error:
			if r != beego.ErrAbort {
				t.Fatal(r)
			}
		default:
			t.Fatal(r)
		}

		if rw.Code != 500 {
			t.Fatal("wrong response code")
		}
	}()

	dc.GetDiagramPage()
}

func TestDiagramController_GetDiagramPage_Success(t *testing.T) {
	rw, ctrl := initBeegoCtrl()

	dc := DiagramController{
		PipelineService: test.MockPipelineService{},
		CodebaseService: test.MockCodebaseService{},
		Controller:      ctrl,
	}

	dc.GetDiagramPage()

	if rw.Code != 200 {
		t.Fatal("wrong response code")
	}
}
