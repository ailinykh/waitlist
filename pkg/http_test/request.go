package httptest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type request struct {
	t       testing.TB
	handler http.Handler

	method  string
	url     string
	data    []byte
	headers map[string]string
}

func (r *request) Request(opts ...func(*request)) *request {
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithMethod(method string) func(*request) {
	return func(request *request) {
		request.method = method
	}
}

func WithUrl(url string) func(*request) {
	return func(request *request) {
		request.url = url
	}
}

func WithHeader(key, value string) func(*request) {
	return func(request *request) {
		request.headers[key] = value
	}
}

func WithData(data []byte) func(*request) {
	return func(request *request) {
		request.data = data
	}
}

func (r *request) ToRespond(opts ...func(*response)) {
	r.t.Helper()

	expected := &response{
		code:        http.StatusOK,
		contentType: "",
		cookies:     make(map[string]*string),
	}
	for _, opt := range opts {
		opt(expected)
	}

	request := httptest.NewRequest(r.method, r.url, bytes.NewReader(r.data))
	response := httptest.NewRecorder()

	for key, value := range r.headers {
		request.Header.Set(key, value)
	}

	r.handler.ServeHTTP(response, request)

	if response.Code != expected.code {
		r.t.Errorf("expected %d but got %d", expected.code, response.Code)
	}

	if len(expected.cookies) > 0 {
		cookies := response.Result().Cookies()
		for k, v := range expected.cookies {
			found := false
			for _, c := range cookies {
				if c.Name == k {
					found = true
					if v != nil && c.Value != *v {
						r.t.Errorf("expected '%s' cookie to be: '%s', got '%s'", k, *v, c.Value)
					}
				}
			}
			if !found {
				r.t.Errorf("expected set cookie: '%s' with value: '%s'", k, *v)
			}
		}
	}

	if len(expected.contentType) > 0 {
		contentType := response.Header().Get("Content-Type")
		if contentType != expected.contentType {
			r.t.Errorf("expected %s but got %s", expected.contentType, contentType)
		}
	}

	if expected.body != nil {
		res := response.Result()
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			r.t.Errorf("unexpected body %s", err)
		}
		if string(expected.body) != string(data) {
			r.t.Errorf("expected %s but got %s", expected.body, data)
		}
	}
}
