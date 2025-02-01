package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIGetEntries(t *testing.T) {
	app, _ := makeSUT(t)

	t.Run("it returns entries from the database", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}
	})
}
