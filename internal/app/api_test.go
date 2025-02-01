package app_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/repository"
)

func TestAPIGetEntries(t *testing.T) {
	repo := repository.New(newDb(t))
	handler := app.NewAPIHandlerFunc(slog.Default(), repo)

	t.Run("it returns entries from the database", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/api", nil)
		response := httptest.NewRecorder()

		handler(response, request)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}
	})
}
