package app

import (
	"log/slog"
	"strings"

	"github.com/ailinykh/waitlist/internal/clock"
)

type Config struct {
	clock               clock.Clock
	port                int
	telegramBotToken    string
	telegramBotEndpoint string
	jwtSecret           string
	staticFilesDir      string
}

func WithClock(clock clock.Clock) func(*Config) {
	return func(c *Config) {
		c.clock = clock
	}
}

func WithPort(port int) func(*Config) {
	return func(c *Config) {
		c.port = port
	}
}

func WithTelegramBotToken(token string) func(*Config) {
	return func(c *Config) {
		c.telegramBotToken = token
	}
}

func WithTelegramBotEndpoint(endpoint string) func(*Config) {
	return func(c *Config) {
		c.telegramBotEndpoint = endpoint
	}
}

func WithJwtSecret(token string) func(*Config) {
	return func(c *Config) {
		c.jwtSecret = token
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
		slog.String("telegramBotToken", safe(c.telegramBotToken)),
		slog.String("jwtSecret", safe(c.jwtSecret)),
		slog.String("staticFilesDir", c.staticFilesDir),
	)
}
