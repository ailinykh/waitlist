package telegram

import (
	"bytes"
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
