package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func NewBot(token, endpoint string, logger *slog.Logger) (*Bot, error) {
	me, err := getMe(endpoint, token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		User:     me,
		client:   http.DefaultClient,
		endpoint: endpoint,
		token:    token,
		l:        logger.With("username", me.Username),
	}, nil
}

type Bot struct {
	*User
	client   *http.Client
	endpoint string
	token    string
	l        *slog.Logger
}

func getMe(endpoint, token string) (*User, error) {
	resp, err := http.DefaultClient.Get(endpoint + "/bot" + token + "/getMe")
	if err != nil {
		return nil, fmt.Errorf("failed to connect Telegram API %w", err)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read `GetMe` body %w", err)
	}

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
		Result      User   `json:"result"`
	}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %w", err)
	}

	if !r.Ok {
		return nil, fmt.Errorf("telegram error: %s", r.Description)
	}

	return &r.Result, nil
}

func (b *Bot) SendMessage(chatID int64, text string) (*Message, error) {
	o := struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: chatID,
		Text:   text,
	}

	req, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed to pack message data %w", err)
	}

	resp, err := b.client.Post(b.endpoint+"/bot"+b.token+"/sendMessage", "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, fmt.Errorf("failed to connect Telegram API %w", err)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read `GetMe` body %w", err)
	}

	var r struct {
		Result Message `json:"result"`
	}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %w", err)
	}

	return &r.Result, nil
}

func (b *Bot) GetUpdates(offset, timeout int64) ([]*Update, error) {
	urlString := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=%d", b.endpoint, b.token, offset, timeout)
	b.l.Debug("start polling...", "offset", offset, "timeout", timeout)
	resp, err := b.client.Get(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Telegram API %w", err)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read `GetMe` body %w", err)
	}

	b.l.Debug("received response", "data", data)

	var r struct {
		Ok          bool      `json:"ok"`
		Description string    `json:"description"`
		Result      []*Update `json:"result"`
	}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %w", err)
	}

	if !r.Ok {
		return nil, fmt.Errorf("telegram error: %s", r.Description)
	}

	return r.Result, nil
}
