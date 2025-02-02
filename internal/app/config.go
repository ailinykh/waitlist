package app

import (
	"log/slog"
	"strings"
)

type Config struct {
	port                   int
	telegramApiSecretToken string
	staticFilesDir         string
}

func WithPort(port int) func(*Config) {
	return func(c *Config) {
		c.port = port
	}
}

func WithTelegramApiSecretToken(token string) func(*Config) {
	return func(c *Config) {
		c.telegramApiSecretToken = token
	}
}

func WithStaticFilesDir(dir string) func(*Config) {
	return func(c *Config) {
		c.staticFilesDir = dir
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
		slog.Int("port", c.port),
		slog.String("telegramApiSecretToken", safe(c.telegramApiSecretToken)),
		slog.String("staticFilesDir", c.staticFilesDir),
	)
}
