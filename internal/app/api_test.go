package app_test

import (
	"net/http"
	"testing"

	h "github.com/ailinykh/waitlist/pkg/http_test"
)

func TestAPIGetEntries(t *testing.T) {
	app, _ := makeSUT(t)

	t.Run("it returns entries from the database", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api"),
		).ToRespond(
			h.WithCode(http.StatusOK),
			h.WithContentType("application/json"),
		)
	})
}
