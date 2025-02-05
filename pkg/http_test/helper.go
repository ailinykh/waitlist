package httptest

import (
	"net/http"
	"testing"
)

func Expect(t testing.TB, handler http.Handler) *request {
	return &request{
		t:       t,
		handler: handler,
		method:  http.MethodGet,
		url:     "/",
		data:    nil,
		headers: make(map[string]string),
	}
}
