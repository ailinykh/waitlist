package app_test

import (
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/clock"
	h "github.com/ailinykh/waitlist/pkg/http_test"
)

func TestLoginAPI(t *testing.T) {
	app, _ := makeSUT(t,
		app.WithJwtSecret("jwt-secret"),
		// RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"
		clock.WithTime(clock.MustParse("2013-08-14T22:00:00.123456789Z")),
	)

	t.Run("it returns telegram oauth data", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api/telegram/oauth"),
		).ToRespond(
			h.WithCode(200),
			h.WithContentType("application/json"),
			h.WithBody([]byte(`{"username":""}`)),
		)
	})
}
