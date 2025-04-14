package cmd

import (
	"context"
	"io"
	"net/http"
)

type MockRESTClient struct {
	PostFunc   func(endpoint string, body interface{}, response interface{}) error
	GetFunc    func(endpoint string, response interface{}) error
	DeleteFunc func(endpoint string, response interface{}) error
	PatchFunc  func(endpoint string, body io.Reader, response interface{}) error
}

// Updated the Post method to match the RESTClientInterface signature
func (m *MockRESTClient) Post(endpoint string, body io.Reader, response interface{}) error {
	if m.PostFunc != nil {
		return m.PostFunc(endpoint, body, response)
	}
	return nil
}

func (m *MockRESTClient) Get(endpoint string, response interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(endpoint, response)
	}
	return nil
}

func (m *MockRESTClient) Delete(endpoint string, response interface{}) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(endpoint, response)
	}
	return nil
}

// Updated the Patch method to match the RESTClientInterface signature
func (m *MockRESTClient) Patch(endpoint string, body io.Reader, response interface{}) error {
	if m.PatchFunc != nil {
		return m.PatchFunc(endpoint, body, response)
	}
	return nil
}

func (m *MockRESTClient) RequestWithContext(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	return nil, nil
}

func (m *MockRESTClient) Request(method string, path string, body io.Reader) (*http.Response, error) {
	return nil, nil
}

func (m *MockRESTClient) DoWithContext(ctx context.Context, method string, path string, body io.Reader, response interface{}) error {
	return nil
}

func (m *MockRESTClient) Do(method string, path string, body io.Reader, response interface{}) error {
	return nil
}

func (m *MockRESTClient) Put(path string, body io.Reader, resp interface{}) error {
	return nil
}
