package app_test

import (
	"log/slog"
	"testing"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/repository"
)

func TestWaitlistSavesUserInTheDatabase(t *testing.T) {
	svr := makeServerMock(t, "test_waitlist")
	repo := repository.New(newDb(t))
	bot, err := telegram.NewBot("Token:1234", svr.URL, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	waitlist := app.NewWaitlist(bot, repo, slog.Default())

	t.Run("it saves user message in the database", func(t *testing.T) {
		err := waitlist.Run(t.Context())
		if err != nil {
			t.Fatalf("failed to run waitlist logic %s", err)
		}

		entries, err := repo.GetAllEntries(t.Context())
		if err != nil {
			t.Fatalf("failed to get all entries %s", err)
		}

		if len(entries) != 1 {
			t.Fatalf("expected single entry but got %d", len(entries))
		}

		if entries[0].Username != "jappleseed" {
			t.Errorf("Unexpected username %s", entries[0].Username)
		}

		if entries[0].Message != "/start" {
			t.Errorf("Unexpected message %s", entries[0].Message)
		}
	})

	t.Run("it saves one more message in the database", func(t *testing.T) {
		err := waitlist.Run(t.Context())
		if err != nil {
			t.Fatalf("failed to run waitlist logic %s", err)
		}

		all, err := repo.GetAllEntries(t.Context())
		if err != nil {
			t.Fatalf("failed to get all entries %s", err)
		}

		if len(all) != 2 {
			t.Errorf("Expected 2 entry, got %d", len(all))
		}
	})
}

func TestWaitlistRespondsToPrivateMessage(t *testing.T) {
	t.Run("it accepts different command formats and responds with message", func(t *testing.T) {
		svr := makeServerMock(t, "test_waitlist")
		repo := repository.New(newDb(t))
		bot, err := telegram.NewBot("Token:1234", svr.URL, slog.Default())
		if err != nil {
			t.Fatal(err)
		}

		waitlist := app.NewWaitlist(bot, repo, slog.Default())
		if err := waitlist.Run(t.Context()); err != nil {
			t.Fatalf("failed to run waitlist logic %s", err)
		}
	})
}
