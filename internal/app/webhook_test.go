package app_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/app"
	"github.com/ailinykh/waitlist/internal/repository"
)

func TestWebhookSavesUserRequestInDatabase(t *testing.T) {
	repo := repository.New(newDb(t))
	handler := app.NewWebhookHandlerFunc(slog.Default(), &telegram.Parser{}, repo)

	t.Run("it saves user message in the database", func(t *testing.T) {
		data := message(t, "chat_private_command_start")
		request, _ := http.NewRequest(http.MethodPost, "/webhook/botusername", bytes.NewReader(data))
		response := httptest.NewRecorder()

		handler(response, request)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}

		entry, err := repo.GetByID(request.Context(), 1)
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
		request, _ := http.NewRequest(http.MethodPost, "/webhook/botusername", bytes.NewReader(requestMessage))
		response, data := perform(t, request, handler)

		if response.Code != 200 {
			t.Errorf("expected 200 but got %d", response.Code)
		}

		all, err := repo.GetAll(request.Context())
		if err != nil {
			t.Errorf("failed to get all entries %s", err)
		}

		if len(all) != 2 {
			t.Errorf("Expected 2 entry, got %d", len(all))
		}

		expected := string(responseMessage)
		actual := string(data)
		if expected != actual {
			t.Errorf("expected %s but got %s", expected, actual)
		}
	})
}

func TestWebhookRespondsToPrivateMessage(t *testing.T) {
	repo := repository.New(newDb(t))
	handler := app.NewWebhookHandlerFunc(slog.Default(), &telegram.Parser{}, repo)

	t.Run("it accepts different command fotmats and send's reply message", func(t *testing.T) {
		testWith := func(t *testing.T, requestMessage []byte, responseMessage []byte) {
			request, _ := http.NewRequest(http.MethodPost, "/webhook/botusername", bytes.NewReader(requestMessage))

			response, data := perform(t, request, handler)

			if response.Code != 200 {
				t.Errorf("expected 200 but got %d", response.Code)
			}

			all, err := repo.GetAll(request.Context())
			if err != nil {
				t.Errorf("failed to get all entries %s", err)
			}

			if len(all) != 0 {
				t.Errorf("Expected 0 entry, got %d", len(all))
			}

			expected := string(responseMessage)
			actual := string(data)
			if expected != actual {
				t.Errorf("expected %s but got %s", expected, actual)
			}

			contentType := response.Header().Get("content-type")
			if contentType != "application/json" {
				t.Errorf(`expected "application/json" but got "%s" Content-Type`, contentType)
			}
		}

		testWith(t, message(t, "chat_private_message_ping"), message(t, "chat_private_message_ping_response"))
		testWith(t, message(t, "chat_private_command_ping"), message(t, "chat_private_command_ping_response"))
	})
}

func perform(t testing.TB, request *http.Request, handler http.Handler) (*httptest.ResponseRecorder, []byte) {
	t.Helper()
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	res := response.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unexpected body %s", err)
	}
	return response, data
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
