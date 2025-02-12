package app_test

import (
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/clock"
	h "github.com/ailinykh/waitlist/pkg/http_test"
)

func TestAPIGetEntries(t *testing.T) {
	app, _ := makeSUT(t,
		app.WithJwtSecret("jwt-secret"),
		// RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"
		clock.WithTime(clock.MustParse("2013-08-14T23:00:00.123456789Z")),
	)

	t.Run("it returns entries from the database", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api"),
			// admin auth token
			h.WithHeader("Cookie", "auth=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXlsb2FkIjp7ImlkIjoxLCJ1c2VyX2lkIjoxMSwiZmlyc3RfbmFtZSI6ImNhdCIsImxhc3RfbmFtZSI6InBlcnNvbiIsInVzZXJuYW1lIjoiaWxvdmVjYXRzIiwicm9sZSI6ImFkbWluIn0sInR0bCI6MTM4NTE1NzYwMH0.QOSWcJf9vU3hAR2bypLxllGmc3yHZaForC18_jxDR0Q"),
		).ToRespond(
			h.WithCode(200),
			h.WithContentType("application/json"),
		)
	})

	t.Run("it requires admin role", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api"),
			// user auth token
			h.WithHeader("Cookie", "auth=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXlsb2FkIjp7ImlkIjoxLCJ1c2VyX2lkIjoxMSwiZmlyc3RfbmFtZSI6ImNhdCIsImxhc3RfbmFtZSI6InBlcnNvbiIsInVzZXJuYW1lIjoiaWxvdmVjYXRzIiwicm9sZSI6InVzZXIifSwidHRsIjoxMzg1MTU3NjAwfQ.-GX5NOeqXMjp0uNCL34z1V64v9UvRZvCE4coae9Ftec"),
		).ToRespond(
			h.WithCode(401),
		)
	})
}
