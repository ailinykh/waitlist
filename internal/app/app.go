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
	"github.com/ailinykh/waitlist/internal/clock"
	"github.com/ailinykh/waitlist/internal/middleware"

	"github.com/ailinykh/waitlist/internal/repository"
)

type Repo interface {
	GetAllEntries(ctx context.Context) ([]repository.Waitlist, error)
	GetEntryByID(ctx context.Context, id uint64) (repository.Waitlist, error)
	CreateEntry(ctx context.Context, arg repository.CreateEntryParams) (sql.Result, error)
	GetUserByUserID(ctx context.Context, userID int64) (repository.User, error)
	CreateUser(ctx context.Context, arg repository.CreateUserParams) (sql.Result, error)
}

func New(logger *slog.Logger, repo Repo, opts ...func(*Config)) App {
	config := &Config{
		clock:                  clock.New(),
		port:                   8080,
		telegramApiSecretToken: "",
		staticFilesDir:         "web/build",
	}

	for _, opt := range opts {
		opt(config)
	}

	logger.Info("creating app", slog.Any("config", config))

	return &appImpl{
		config: config,
		logger: logger,
		repo:   repo,
		stack:  newStack(logger, config, repo),
	}
}

type App interface {
	http.Handler
	Run(context.Context) error
}

type appImpl struct {
	config *Config
	logger *slog.Logger
	repo   Repo
	stack  http.Handler
}

func (app *appImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.stack.ServeHTTP(w, r)
}

func (app *appImpl) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", app.config.port)
	server := http.Server{
		Addr:    addr,
		Handler: app,
	}

	done := make(chan struct{})
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("failed to listen and serve", slog.Any("error", err))
		}
		close(done)
	}()

	app.logger.Info("server listening", slog.String("addr", addr))
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

func newStack(logger *slog.Logger, config *Config, repo Repo) http.Handler {
	router := http.NewServeMux()

	fs := http.Dir(config.staticFilesDir)
	router.Handle("/",
		middleware.NewSPA(middleware.ServeFileContents("index.html", fs))(
			http.FileServer(fs),
		),
	)

	router.HandleFunc("GET /api/telegram/oauth", NewOAuthHandlerFunc(config, logger))
	router.HandleFunc("GET /api/telegram/oauth/token", NewCallbackHandlerFunc(config, repo, config.clock, logger))

	router.Handle(
		"POST /webhook/{bot}",
		middleware.HeaderAuth("X-Telegram-Bot-Api-Secret-Token", config.telegramApiSecretToken, logger)(
			NewWebhookHandlerFunc(logger, &telegram.Parser{}, repo),
		),
	)

	authStack := middleware.CreateStack(
		middleware.JwtAuth(config.jwtSecret, middleware.User{}, config.clock, logger),
		middleware.RoleAuth("admin", logger),
	)

	router.Handle("GET /api/entries", authStack(NewAPIHandlerFunc(logger, repo)))

	stack := middleware.CreateStack(
		middleware.Logging(logger),
	)

	return stack(router)
}
