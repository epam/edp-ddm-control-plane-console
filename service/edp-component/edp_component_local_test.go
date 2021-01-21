package edpcomponent

import "testing"

func TestLocal_GetEDPComponent(t *testing.T) {
	localEDPComponent := MakeLocalLinks(map[string]string{
		"foo": "bar",
	})

	com, err := localEDPComponent.GetEDPComponent("foo")
	if err != nil {
		t.Fatal(err)
	}

	if com.URL != "bar" {
		t.Fatal("wrong url returned")
	}
}

func TestLocal_GetEDPComponent_Failure(t *testing.T) {
	local := MakeLocalLinks(map[string]string{"foo": "bar"})

	_, err := local.GetEDPComponent("bar")
	if err == nil {
		t.Fatal("no error on wrong component name")
	}

	switch err.(type) {
	case NotFound:
		return
	default:
		t.Fatal("wrong error type returned")
	}
}
