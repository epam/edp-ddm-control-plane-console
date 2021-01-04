package test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	GetRateLimiter() flowcontrol.RateLimiter
	Verb(verb string) *rest.Request
	Post() *rest.Request
	Put() *rest.Request
	Patch(pt types.PatchType) *rest.Request
	Get() *rest.Request
	Delete() *rest.Request
	APIVersion() schema.GroupVersion
}

type MockRestInterface struct {
	PostResponse   *rest.Request
	PutResponse    *rest.Request
	PatchResponse  *rest.Request
	GetResponse    *rest.Request
	DeleteResponse *rest.Request
}

func (MockRestInterface) GetRateLimiter() flowcontrol.RateLimiter {
	panic("not implemented")
}

func (MockRestInterface) Verb(verb string) *rest.Request {
	panic("not implemented")
}

func (m MockRestInterface) Post() *rest.Request {
	return m.PostResponse
}

func (m MockRestInterface) Put() *rest.Request {
	return m.PutResponse
}

func (m MockRestInterface) Patch(pt types.PatchType) *rest.Request {
	return m.PatchResponse
}

func (m MockRestInterface) Get() *rest.Request {
	return m.GetResponse
}

func (m MockRestInterface) Delete() *rest.Request {
	return m.DeleteResponse
}

func (MockRestInterface) APIVersion() schema.GroupVersion {
	panic("not implemented")
}

type MockBackoffManager struct{}

func (m MockBackoffManager) UpdateBackoff(actualURL *url.URL, err error, responseCode int) {}

func (m MockBackoffManager) CalculateBackoff(actualURL *url.URL) time.Duration {
	return time.Millisecond
}

func (m MockBackoffManager) Sleep(d time.Duration) {}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockHTTPClient struct {
	DoResponse *http.Response
	DoError    error
}

func (m MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoResponse, m.DoError
}

type MockJSONEncoderDecoder struct{}

func (m MockJSONEncoderDecoder) Decode(data []byte, defaults *schema.GroupVersionKind,
	into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	if err := json.Unmarshal(data, into); err != nil {
		return nil, nil, err
	}

	return into, nil, nil
}

func (m MockJSONEncoderDecoder) Encode(obj runtime.Object, w io.Writer) error {
	return json.NewEncoder(w).Encode(obj)
}

func NewRequestShortcut(httpClient rest.HTTPClient) *rest.Request {
	return rest.NewRequest(httpClient, "", nil, "", rest.ContentConfig{},
		rest.Serializers{
			RenegotiatedDecoder: func(contentType string, params map[string]string) (runtime.Decoder, error) {
				return MockJSONEncoderDecoder{}, nil
			},
			Encoder: MockJSONEncoderDecoder{},
		}, &rest.NoBackoff{}, flowcontrol.NewFakeAlwaysRateLimiter(), time.Second)
}

func NewResponseShortcut(responseBody string) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(responseBody)))
	header := make(http.Header)
	header.Add("Content-Type", "application/json")

	return &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Body:       r,
		StatusCode: http.StatusOK,
		//Request:    req,
	}, nil
}
