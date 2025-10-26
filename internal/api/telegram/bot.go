package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func NewBot(token, endpoint string) *Bot {
	return &Bot{
		client:   http.DefaultClient,
		endpoint: endpoint,
		token:    token,
	}
}

type Bot struct {
	client   *http.Client
	endpoint string
	token    string
}

func (b *Bot) GetMe() (*User, error) {
	resp, err := b.client.Get(b.endpoint + "/bot" + b.token + "/getMe")
	if err != nil {
		return nil, fmt.Errorf("failed to connect Telegram API %w", err)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read `GetMe` body %w", err)
	}

	var r struct {
		Result User `json:"result"`
	}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %w", err)
	}

	return &r.Result, nil
}
