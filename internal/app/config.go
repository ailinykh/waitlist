package app

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
