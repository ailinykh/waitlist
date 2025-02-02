package app

import (
	"log/slog"
	"strings"
)

type Config struct {
	Port                   int
	TelegramApiSecretToken string
	StaticFilesDir         string
}

func WithPort(port int) func(*Config) {
	return func(c *Config) {
		c.Port = port
	}
}

func WithTelegramApiSecretToken(token string) func(*Config) {
	return func(c *Config) {
		c.TelegramApiSecretToken = token
	}
}

func WithStaticFilesDir(dir string) func(*Config) {
	return func(c *Config) {
		c.StaticFilesDir = dir
	}
}

func (c Config) LogValue() slog.Value {
	safe := func(text string) string {
		if len(text) < 5 {
			return strings.Repeat("*", len(text))
		}
		return text[:2] + strings.Repeat("*", len(text)-4) + text[len(text)-2:]
	}
	return slog.GroupValue(
		slog.Int("port", c.Port),
		slog.String("bot_api_secret_token", safe(c.TelegramApiSecretToken)),
		slog.String("static_files_dir", c.StaticFilesDir),
	)
}
