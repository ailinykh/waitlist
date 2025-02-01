package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/database"
	"github.com/ailinykh/waitlist/internal/repository"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := slog.Default()
	repo := repository.New(db(logger))

	app := app.New(
		logger,
		repo,
		app.WithTelegramApiSecretToken(os.Getenv("TELEGRAM_BOT_API_SECRET_TOKEN")),
	)
	if err := app.Run(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	logger.Info("shoutdown app...")
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
