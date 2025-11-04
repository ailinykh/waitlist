package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := NewLogger()
	repo := repository.New(db(logger))

	app := app.New(
		logger,
		repo,
		app.WithTelegramApiSecretToken(os.Getenv("TELEGRAM_BOT_API_SECRET_TOKEN")),
		app.WithTelegramBotToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
		app.WithJwtSecret(os.Getenv("JWT_SECRET")),
	)
	if err := app.Run(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	logger.Info("shutdown app...")
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
