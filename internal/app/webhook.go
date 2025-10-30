package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/repository"
)

type TelegramUpdateContextKey struct{}

func parseUpdate(r *http.Request, _ *slog.Logger, parser *telegram.Parser) (*telegram.Update, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("no body passed")
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return parser.Parse(data)
}

func handleUpdate(r *http.Request, logger *slog.Logger, repo Repo, bot *telegram.Bot) ([]byte, error) {
	update := r.Context().Value(TelegramUpdateContextKey{}).(*telegram.Update)
	if update == nil {
		return nil, fmt.Errorf("no update in request context")
	}

	if update.Message == nil {
		logger.Info("ignoring non-message update", slog.Int("id", update.ID))
		return nil, nil
	}

	if strings.TrimPrefix(update.Message.Text, "/") == "ping" {
		logger.Info("ping message received", slog.Int("id", update.ID))
		_, err := bot.SendMessage(update.Message.Chat.ID, "pong")
		return nil, err
	}

	arg := repository.CreateEntryParams{
		UserID:      update.Message.From.ID,
		FirstName:   update.Message.From.FirstName,
		LastName:    update.Message.From.LastName,
		Username:    update.Message.From.Username,
		Message:     update.Message.Text,
		BotUsername: r.PathValue("bot"),
	}

	if _, err := repo.CreateEntry(r.Context(), arg); err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	if strings.HasPrefix(update.Message.Text, "/start") {
		_, err := bot.SendMessage(update.Message.Chat.ID, "This bot is not available in your region yet. Please come back later.")
		return nil, err
	}

	return nil, nil
}

func NewWebhookHandlerFunc(logger *slog.Logger, parser *telegram.Parser, bot *telegram.Bot, repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Error("only POST method allowed", slog.String("method", r.Method))
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		update, err := parseUpdate(r, logger, parser)
		if err != nil {
			logger.Error("failed to parse upadte request", slog.Any("error", err))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TelegramUpdateContextKey{}, update)
		req := r.WithContext(ctx)
		data, err := handleUpdate(req, logger, repo, bot)
		if err != nil {
			logger.Error("failed to handle request", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
