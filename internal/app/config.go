package app

import (
	"log/slog"
	"strings"

	"github.com/ailinykh/waitlist/internal/clock"
)

type Config struct {
	clock                  clock.Clock
	port                   int
	telegramApiSecretToken string
	telegramBotToken       string
	jwtSecret              string
	staticFilesDir         string
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

func WithTelegramApiSecretToken(token string) func(*Config) {
	return func(c *Config) {
		c.telegramApiSecretToken = token
	}
}

func WithTelegramBotToken(token string) func(*Config) {
	return func(c *Config) {
		c.telegramBotToken = token
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
		slog.String("telegramApiSecretToken", safe(c.telegramApiSecretToken)),
		slog.String("telegramBotToken", safe(c.telegramBotToken)),
		slog.String("jwtSecret", safe(c.jwtSecret)),
		slog.String("staticFilesDir", c.staticFilesDir),
	)
}
