package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"
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

func New(logger *slog.Logger, repo Repo, clock clock.Clock, templates fs.FS, opts ...func(*Config)) App {
	config := &Config{
		port:                   8080,
		telegramApiSecretToken: "",
		staticFilesDir:         "web",
	}

	for _, opt := range opts {
		opt(config)
	}

	logger.Info("creating app", slog.Any("config", config))

	tmpl := template.Must(template.New("").ParseFS(templates, "templates/*"))

	return &appImpl{
		config: config,
		logger: logger,
		repo:   repo,
		stack:  newStack(logger, config, repo, clock, tmpl),
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

func NewIndexHandlerFunc(repo Repo, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			tmpl.ExecuteTemplate(w, "error.html", struct {
				Title       string
				Description string
			}{
				Title:       "Page Not Found",
				Description: "Sorry, we couldn’t find the page you’re looking for.",
			})
			return
		}

		entries, _ := repo.GetAllEntries(r.Context())

		tmpl.ExecuteTemplate(w, "index.html", struct {
			Entries []repository.Waitlist
			Total   int
		}{
			Entries: entries,
			Total:   len(entries),
		})
	}
}

func newStack(logger *slog.Logger, config *Config, repo Repo, clock clock.Clock, tmpl *template.Template) http.Handler {
	router := http.NewServeMux()

	// fs := http.FileServer(http.Dir(app.config.StaticFilesDir))
	// router.Handle("/", fs)

	bot, err := telegram.GetMe(config.telegramBotToken)
	if err != nil {
		panic(err)
	}

	router.HandleFunc("GET /login", NewLoginHandlerFunc(bot.Username, tmpl))
	router.HandleFunc("GET /logout", NewLogutHandlerFunc())

	router.HandleFunc("GET /api/telegram/callback", NewCallbackHandlerFunc(config, repo, clock, logger))

	router.Handle(
		"POST /webhook/{bot}",
		middleware.HeaderAuth("X-Telegram-Bot-Api-Secret-Token", config.telegramApiSecretToken, logger)(
			NewWebhookHandlerFunc(logger, &telegram.Parser{}, repo),
		),
	)

	authStack := middleware.CreateStack(
		middleware.JwtAuth(config.jwtSecret, middleware.User{}, clock, logger),
		middleware.RoleAuth("admin", logger),
	)

	router.HandleFunc("/", http.NotFound)
	router.Handle("GET /", authStack(NewIndexHandlerFunc(repo, tmpl)))
	router.Handle("GET /api", authStack(NewAPIHandlerFunc(logger, repo)))

	stack := middleware.CreateStack(
		middleware.Logging(logger),
	)

	return stack(router)
}
