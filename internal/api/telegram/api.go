package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetMe(token string) (*User, error) {
	resp, err := http.DefaultClient.Get("https://api.telegram.org/bot" + token + "/getMe")
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
