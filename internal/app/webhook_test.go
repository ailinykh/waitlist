package app_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/repository"
	h "github.com/ailinykh/waitlist/pkg/http_test"
)

func TestWebhookAllowedHttpMethods(t *testing.T) {
	// should test handler instead of `app` to avoid 404 page on GET request
	repo := repository.New(newDb(t))
	handler := app.NewWebhookHandlerFunc(slog.Default(), &telegram.Parser{}, repo)

	testHttpMethod := func(t *testing.T, method string, status int, body []byte) {
		t.Run(fmt.Sprintf("it responds with %d status to %s request", status, method), func(t *testing.T) {
			h.Expect(t, handler).Request(
				h.WithUrl("/webhook/botusername"),
				h.WithMethod(method),
				h.WithData(body),
			).ToRespond(
				h.WithCode(status),
			)
		})
	}

	testHttpMethod(t, http.MethodGet, 405, nil)
	testHttpMethod(t, http.MethodPatch, 405, nil)
	testHttpMethod(t, http.MethodPut, 405, nil)
	testHttpMethod(t, http.MethodDelete, 405, nil)
	testHttpMethod(t, http.MethodPost, 400, []byte(""))
	testHttpMethod(t, http.MethodPost, 200, []byte("{}"))
}

func TestWebhookSavesUserInTheDatabase(t *testing.T) {
	app, repo := makeSUT(t)

	t.Run("it saves user message in the database", func(t *testing.T) {
		data := message(t, "chat_private_command_start")
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData(data),
		).ToRespond(
			h.WithCode(http.StatusOK),
		)

		entry, err := repo.GetEntryByID(context.TODO(), 1)
		if err != nil {
			t.Errorf("failed to get all entries %s", err)
		}

		if entry.Username != "jappleseed" {
			t.Errorf("Unexpected username %s", entry.Username)
		}

		if entry.Message != "/start" {
			t.Errorf("Unexpected message %s", entry.Message)
		}
	})

	t.Run("it saves one more message in the database", func(t *testing.T) {
		requestMessage := message(t, "chat_private_command_start")
		responseMessage := message(t, "chat_private_command_start_response")
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData(requestMessage),
		).ToRespond(
			h.WithCode(http.StatusOK),
			h.WithBody(responseMessage),
		)

		all, err := repo.GetAllEntries(context.TODO())
		if err != nil {
			t.Errorf("failed to get all entries %s", err)
		}

		if len(all) != 2 {
			t.Errorf("Expected 2 entry, got %d", len(all))
		}
	})
}

func TestWebhookRespondsToPrivateMessage(t *testing.T) {
	app, _ := makeSUT(t)
	t.Run("it accepts different command formats and responds with message", func(t *testing.T) {
		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData(message(t, "chat_private_message_ping")),
		).ToRespond(
			h.WithCode(http.StatusOK),
			h.WithBody(message(t, "chat_private_message_ping_response")),
		)

		h.Expect(t, app).Request(
			h.WithUrl("/webhook/botusername"),
			h.WithMethod(http.MethodPost),
			h.WithData(message(t, "chat_private_command_ping")),
		).ToRespond(
			h.WithCode(http.StatusOK),
			h.WithBody(message(t, "chat_private_command_ping_response")),
		)
	})
}

func message(t testing.TB, filename string) []byte {
	t.Helper()
	filePath := path.Join(cwd(t), "test", "fixtures", filename+".json")
	t.Logf("using fixture %s", filePath)

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Error(err)
	}

	return bytes
}
