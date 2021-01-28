package edpcomponent

import (
	"ddm-admin-console/test"
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestServiceK8S_Get(t *testing.T) {
	getResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"type": "bar"}}`)

	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    nil,
	}

	svc := MakeServiceK8S(test.MockRestInterface{
		GetResponse: test.NewRequestShortcut(getHTTPClient),
	})

	item, err := svc.Get("mike-test", "test")
	if err != nil {
		t.Fatal(err)
	}

	if item.Name != "test" {
		t.Fatal("wrong component name")
	}
}

func TestServiceK8S_GetFailure(t *testing.T) {
	mockErr := errors.New("k8s fatal")
	getResponse, _ := test.NewResponseShortcut(
		`{"metadata": {"name": "test"}, "spec": {"type": "bar"}}`)

	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    mockErr,
	}

	svc := MakeServiceK8S(test.MockRestInterface{
		GetResponse: test.NewRequestShortcut(getHTTPClient),
	})

	_, err := svc.Get("mike-test", "test")
	if err == nil {
		t.Fatal("no error on k8s fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Log(fmt.Sprintf("%+v", err))
		t.Fatal("wrong error returned")
	}
}

func TestServiceK8S_GetAll(t *testing.T) {
	getResponse, _ := test.NewResponseShortcut(
		`{"items": [{"metadata": {"name": "test"}, "spec": {"type": "bar"}}]}`)

	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    nil,
	}

	svc := MakeServiceK8S(test.MockRestInterface{
		GetResponse: test.NewRequestShortcut(getHTTPClient),
	})

	items, err := svc.GetAll("mike-test")
	if err != nil {
		t.Fatal(err)
	}

	if len(items) == 0 {
		t.Fatal("no items returned")
	}
}

func TestServiceK8S_GetAllFailure(t *testing.T) {
	mockErr := errors.New("k8s fatal")
	getResponse, _ := test.NewResponseShortcut(
		`{"items": [{"metadata": {"name": "test"}, "spec": {"type": "bar"}}]}`)

	getHTTPClient := test.MockHTTPClient{
		DoResponse: getResponse,
		DoError:    mockErr,
	}

	svc := MakeServiceK8S(test.MockRestInterface{
		GetResponse: test.NewRequestShortcut(getHTTPClient),
	})

	_, err := svc.GetAll("mike-test")
	if err == nil {
		t.Fatal("no error on k8s fatal")
	}

	if errors.Cause(err) != mockErr {
		t.Log(fmt.Sprintf("%+v", err))
		t.Fatal("wrong error returned")
	}
}
