package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/middleware"
	"github.com/ailinykh/waitlist/internal/repository"
)

type Repo interface {
	GetAll(ctx context.Context) ([]repository.Waitlist, error)
	GetByID(ctx context.Context, id uint64) (repository.Waitlist, error)
	CreateEntry(ctx context.Context, arg repository.CreateEntryParams) (sql.Result, error)
}

func New(logger *slog.Logger, repo Repo, opts ...func(*Config)) App {
	config := &Config{
		Port:                   8080,
		TelegramApiSecretToken: "",
		StaticFilesDir:         "web",
	}

	for _, opt := range opts {
		opt(config)
	}

	return &appImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

type App interface {
	Run(context.Context) error
	Handler() http.Handler
}

type appImpl struct {
	config *Config
	logger *slog.Logger
	repo   Repo
}

func (app *appImpl) Handler() http.Handler {
	router := http.NewServeMux()

	// fs := http.FileServer(http.Dir(app.config.StaticFilesDir))
	// router.Handle("/", fs)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<div style="text-align: center; font-family: 'Roboto', -apple-system, 'Helvetica Neue', sans-serif;">
			<H1>Welcome to the waitlist!</H1>
		</div>
	</body>
</html>
`))
	})
	router.HandleFunc("GET /api", NewAPIHandlerFunc(app.logger, app.repo))
	router.HandleFunc("/webhook/{bot}", NewWebhookHandlerFunc(app.logger, &telegram.Parser{}, app.repo))

	stack := middleware.CreateStack(
		middleware.NewLogging(app.logger),
	)

	if len(app.config.TelegramApiSecretToken) > 0 {
		stack = middleware.CreateStack(
			stack,
			middleware.NewAuth(app.config.TelegramApiSecretToken, app.logger),
		)
	}

	return stack(router)
}

func (app *appImpl) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", app.config.Port)
	server := http.Server{
		Addr:    addr,
		Handler: app.Handler(),
	}

	done := make(chan struct{})
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("failed to listen and serve", slog.Any("error", err))
		}
		close(done)
	}()

	app.logger.Info("Server listening", slog.String("addr", addr))
	select {
	case <-done:
		break
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		server.Shutdown(ctx)
		cancel()
	}
	return nil
}
