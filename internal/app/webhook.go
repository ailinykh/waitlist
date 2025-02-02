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

type ContextKey = string

type TelegramUpdate struct{}

func parseUpdate(r *http.Request, _ *slog.Logger, parser *telegram.Parser) (*telegram.Update, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("no body passed")
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %s", err)
	}

	return parser.Parse(data)
}

func handleUpdate(r *http.Request, logger *slog.Logger, repo Repo) ([]byte, error) {
	update := r.Context().Value(TelegramUpdate{}).(*telegram.Update)
	if update == nil {
		return nil, fmt.Errorf("no update in request context")
	}

	if update.Message == nil {
		logger.Info("ignoring non-message update", slog.Int("id", update.ID))
		return nil, nil
	}

	if strings.TrimPrefix(update.Message.Text, "/") == "ping" {
		logger.Info("ping message received", slog.Int("id", update.ID))
		return telegram.NewResponse("sendMessage",
			telegram.WithChatID(update.Message.Chat.ID),
			telegram.WithText("pong"),
		).ToJSON()
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
		return nil, fmt.Errorf("failed to create entry: %s", err)
	}

	if strings.HasPrefix(update.Message.Text, "/start") {
		return telegram.NewResponse("sendMessage",
			telegram.WithChatID(update.Message.Chat.ID),
			telegram.WithText("This bot is not available in your region yet. Please come back later."),
		).ToJSON()
	}

	return nil, nil
}

func NewWebhookHandlerFunc(logger *slog.Logger, parser *telegram.Parser, repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			logger.Error("only POST method allowed", slog.String("method", r.Method))
			return
		}

		update, err := parseUpdate(r, logger, parser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Error("failed to parse upadte request", slog.Any("error", err))
			return
		}

		ctx := context.WithValue(r.Context(), TelegramUpdate{}, update)
		req := r.WithContext(ctx)
		data, err := handleUpdate(req, logger, repo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("failed to handle request", slog.Any("error", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
