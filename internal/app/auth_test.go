package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/clock"
	h "github.com/ailinykh/waitlist/pkg/http_test"
)

func TestLoginAPI(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot0123456789:TeLeGRAMm_bot-T0keN/getMe" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":{"username":"waitlist_bot"}}`))
	}))
	defer svr.Close()

	app, _ := makeSUT(t,
		app.WithJwtSecret("jwt-secret"),
		app.WithTelegramBotToken("0123456789:TeLeGRAMm_bot-T0keN"),
		app.WithTelegramBotEndpoint(svr.URL),
		// RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"
		app.WithClock(
			clock.New(clock.WithTime(clock.MustParse("2013-08-14T22:00:00.123456789Z"))),
		),
	)

	t.Run("it returns telegram oauth data", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api/telegram/oauth"),
		).ToRespond(
			h.WithCode(200),
			h.WithContentType("application/json"),
			h.WithBody([]byte(`{"username":"waitlist_bot"}`)),
		)
	})
}
