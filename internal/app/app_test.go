package app_test

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestAuthHeader(t *testing.T) {
	app, _ := makeSUT(t, app.WithTelegramApiSecretToken("secret"))

	t.Run("it blocks unauthorized requests", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/webhook/botusername", strings.NewReader("{}"))
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		if response.Code != 403 {
			t.Errorf("expected 403 but got %d", response.Code)
		}
	})

	t.Run("it accepts authorized requests", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/webhook/botusername", strings.NewReader("{}"))
		request.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}
	})
}

func TestAppFrontend(t *testing.T) {
	app, _ := makeSUT(t) // app.WithStaticFilesDir(filepath.Join(cwd(t), "web")),

	t.Run("it returns HTML page", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}

		contentType := response.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("expected text/html; charset=utf-8 but got %s", contentType)
		}
	})
}

func cwd(t testing.TB) string {
	t.Helper()
	wd, err := os.Getwd()

	if err != nil {
		t.Error(err)
	}

	for !strings.HasSuffix(wd, "waitlist") {
		wd = filepath.Dir(wd)
	}
	return wd
}

func newDb(t testing.TB) *sql.DB {
	t.Helper()
	ctx := context.TODO()
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase("waitlist"),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(mysqlContainer); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		t.Errorf("failed to start container: %s", err)
	}

	connectionString := mysqlContainer.MustConnectionString(ctx, "parseTime=true")
	db, err := database.New(slog.Default(),
		database.WithURL("mysql://"+connectionString),
		database.WithMigrations(os.DirFS(cwd(t))),
	)
	if err != nil {
		t.Errorf("failed to open MYSQL connection: %s", err)
	}
	return db
}

func makeSUT(t testing.TB, opts ...func(*app.Config)) (app.App, app.Repo) {
	t.Helper()
	repo := repository.New(newDb(t))
	app := app.New(slog.Default(), repo, opts...)
	return app, repo
}
