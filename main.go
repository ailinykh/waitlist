package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := NewLogger()
	repo := repository.New(db(logger))

	server, err := app.New(
		logger,
		repo,
		app.WithTelegramBotToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
		app.WithJwtSecret(os.Getenv("JWT_SECRET")),
	)

	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	wg.Go(func() {
		if err := server.Run(ctx); err != nil {
			panic(err)
		}
	})

	for _, t := range parseTokens() {
		wg.Go(func() {
			bot, err := telegram.NewBot(t, "https://api.telegram.org", logger)
			if err != nil {
				logger.Error("failed to create waitlist", "error", err)
				return
			}

			waitlist := app.NewWaitlist(bot, repo, logger.With("username", bot.Username))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					err = waitlist.Run(ctx)
					if err != nil {
						logger.Error("failed to run", "username", bot.Username, "error", err)
						return
					}
				}
			}
		})
	}

	<-ctx.Done()
	logger.Info("attempt to shutdown gracefully...")

	wg.Wait()
}

//go:embed migrations/*.sql
var migrations embed.FS

func db(logger *slog.Logger) *sql.DB {
	db, err := database.New(logger,
		database.WithURL(os.Getenv("DATABASE_URL")),
		database.WithMigrations(migrations),
	)
	if err != nil {
		panic(err)
	}
	return db
}

func parseTokens() []string {
	tokens := []string{}
	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx > 0 {
			if strings.HasPrefix(env[:idx], "TELEGRAM_BOT_TOKEN") {
				tokens = append(tokens, env[idx+1:])
			}
		}
	}
	return tokens
}

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   true,
		ReplaceAttr: replaceAttr,
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
		source.Function = filepath.Base(source.Function)
	}
	return a
}
