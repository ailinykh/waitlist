package app_test

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
	h "github.com/ailinykh/waitlist/pkg/http_test"
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestAuthMiddleware(t *testing.T) {
	app, _ := makeSUT(t, app.WithTelegramApiSecretToken("secret"))

	t.Run("it blocks unauthorized requests", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData([]byte("{}")),
		).ToRespond(
			h.WithCode(http.StatusForbidden),
		)
	})

	t.Run("it accepts authorized requests", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithHeader("X-Telegram-Bot-Api-Secret-Token", "secret"),
			h.WithData([]byte("{}")),
		).ToRespond(
			h.WithCode(http.StatusOK),
		)
	})

	t.Run("it pases auth middleware for non-webhook requests", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/"),
		).ToRespond(
			h.WithCode(http.StatusOK),
		)
	})
}

func TestAppFrontend(t *testing.T) {
	app, _ := makeSUT(t) // app.WithStaticFilesDir(filepath.Join(cwd(t), "web")),

	t.Run("it displays HTML page", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/"),
		).ToRespond(
			h.WithCode(200),
			h.WithContentType("text/html; charset=utf-8"),
		)
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
