package telegram

import "encoding/json"

func NewResponse(method string, opts ...func(*Response)) *Response {
	resp := &Response{
		Method: method,
	}
	for _, opt := range opts {
		opt(resp)
	}
	return resp
}

type Response struct {
	Method string `json:"method"`
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text,omitempty"`
}

func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(&r)
}

func WithChatID(chatID int64) func(*Response) {
	return func(r *Response) {
		r.ChatID = chatID
	}
}

func WithText(text string) func(*Response) {
	return func(r *Response) {
		r.Text = text
	}
}
