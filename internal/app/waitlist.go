package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/repository"
)

func NewWaitlist(bot *telegram.Bot, repo Repo, logger *slog.Logger) *Waitlist {
	return &Waitlist{
		bot:    bot,
		offset: 0,
		repo:   repo,
		l:      logger,
	}
}

type Waitlist struct {
	bot    *telegram.Bot
	offset int64
	repo   Repo
	l      *slog.Logger
}

func (w *Waitlist) Run(ctx context.Context) error {
	updates, err := w.bot.GetUpdates(w.offset, 100)
	if err != nil {
		w.l.Error("failed to get updates", "error", err)
		return err
	}

	w.l.Info("got updates", "count", len(updates))

	for _, u := range updates {
		w.offset = u.ID + 1

		if u.Message == nil {
			w.l.Info("ignoring non-message update", "id", u.ID)
			continue
		}

		if strings.TrimPrefix(u.Message.Text, "/") == "ping" {
			w.l.Info("ping message received", "id", u.ID)
			_, err := w.bot.SendMessage(u.Message.Chat.ID, "pong")
			if err != nil {
				w.l.Error("failed to send message", "error", err)
				return err
			}
		}

		arg := repository.CreateEntryParams{
			UserID:      u.Message.From.ID,
			FirstName:   u.Message.From.FirstName,
			LastName:    u.Message.From.LastName,
			Username:    u.Message.From.Username,
			Message:     u.Message.Text,
			BotUsername: w.bot.Username,
		}

		if _, err := w.repo.CreateEntry(ctx, arg); err != nil {
			w.l.Error("failed to create entry", "error", err)
			return err
		}

		if strings.HasPrefix(u.Message.Text, "/start") {
			_, err := w.bot.SendMessage(u.Message.Chat.ID, "This bot is not available in your region yet. Please come back later.")
			if err != nil {
				w.l.Error("failed to send message", "error", err)
				return err
			}
		}
	}
	return nil
}
