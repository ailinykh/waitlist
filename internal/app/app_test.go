package app_test

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/clock"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
	h "github.com/ailinykh/waitlist/pkg/http_test"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestXTokenAuthorizationLogic(t *testing.T) {
	app, _ := makeSUT(t, app.WithTelegramApiSecretToken("secret"))

	t.Run("it blocks x-token unauthorized requests", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData([]byte("{}")),
		).ToRespond(
			h.WithCode(403),
		)
	})

	t.Run("it accepts x-token authorized requests", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithHeader("X-Telegram-Bot-Api-Secret-Token", "secret"),
			h.WithData([]byte("{}")),
		).ToRespond(
			h.WithCode(200),
		)
	})
}

func TestJWTAuthorizationLogic(t *testing.T) {
	app, _ := makeSUT(t,
		app.WithJwtSecret("jwt-secret"),
		app.WithTelegramBotToken("telegram-secret"),
		app.WithStaticFilesDir(filepath.Join(cwd(t), "web/build")),
		// RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"
		app.WithClock(
			clock.New(clock.WithTime(clock.MustParse("2013-08-14T22:00:00.123456789Z"))),
		),
	)

	t.Run("callback creates jwt token", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api/telegram/oauth/token?id=11&first_name=cat&last_name=person&username=ilovecats&photo_url=https%3A%2F%2Ft.me%2Fi%2Fuserpic%2F320%2Floh66&auth_date=1739115445&hash=1ff1e59e43a480fdc802bc0b42e3e68e80ce113ef099b459ee689a9e8a2870ca"),
		).ToRespond(
			h.WithCode(200),
			h.WithBody([]byte(`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXlsb2FkIjp7InVzZXJfaWQiOjExLCJmaXJzdF9uYW1lIjoiY2F0IiwibGFzdF9uYW1lIjoicGVyc29uIiwidXNlcm5hbWUiOiJpbG92ZWNhdHMiLCJyb2xlIjoidXNlciJ9LCJ0dGwiOjEzODUxNTc2MDB9.t76NSZmkry5vQcmDXNQgA2aw5KO40g3N07OSNKa2PkI"}`)),
		)
	})

	t.Run("callback validates url query hash", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/api/telegram/oauth/token?id=12&first_name=cat&last_name=person&username=ilovecats&photo_url=https%3A%2F%2Ft.me%2Fi%2Fuserpic%2F320%2Floh66&auth_date=1739115445&hash=1ff1e59e43a480fdc802bc0b42e3e68e80ce113ef099b459ee689a9e8a2870ca"),
		).ToRespond(
			h.WithCode(400),
		)
	})
}

func TestAppFrontend(t *testing.T) {
	app, _ := makeSUT(t, app.WithStaticFilesDir(filepath.Join(cwd(t), "web/build")))

	t.Run("it does not require authorization for spa", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/"),
		).ToRespond(
			h.WithCode(200),
			h.WithContentType("text/html; charset=utf-8"),
		)
	})

	t.Run("it responds with 404 page", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/non-existent"),
		).ToRespond(
			h.WithCode(404),
		)
	})
}

func cwd(t testing.TB) string {
	t.Helper()
	wd, err := os.Getwd()

	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(filepath.Join(wd, "go.mod"))
	for os.IsNotExist(err) {
		wd = filepath.Dir(wd)
		_, err = os.Stat(filepath.Join(wd, "go.mod"))
	}
	return wd
}

func newDb(t testing.TB) *sql.DB {
	t.Helper()
	postgresContainer, err := postgres.Run(t.Context(),
		"postgres:18-alpine",
		postgres.WithDatabase("waitlist"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connectionString := postgresContainer.MustConnectionString(t.Context(), "sslmode=disable")

	db, err := database.New(slog.Default(),
		database.WithURL(connectionString),
		database.WithMigrations(os.DirFS(cwd(t))),
	)
	if err != nil {
		t.Fatalf("failed to open postgres connection: %s", err)
	}
	return db
}

func makeSUT(t testing.TB, opts ...func(*app.Config)) (app.App, app.Repo) {
	t.Helper()
	repo := repository.New(newDb(t))
	app := app.New(slog.Default(), repo, opts...)
	return app, repo
}

type ResponseMock struct {
	Path string
	Body string
}

func makeServer(t testing.TB, responses []ResponseMock) *httptest.Server {
	requests := []string{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(requests) >= len(responses) {
			t.Errorf("too many unexpected requests")
		}

		response := responses[len(requests)]
		if r.URL.Path != response.Path {
			t.Errorf("expected URL path: '%s', but got '%s'", response.Path, r.URL.Path)
		}

		requests = append(requests, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Body))
	}))
}
