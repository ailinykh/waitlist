package telegram

import (
	"encoding/json"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct{}

func (Parser) Parse(data []byte) (update *Update, err error) {
	err = json.Unmarshal(data, &update)
	if err != nil {
		return nil, err
	}
	return update, nil
}
