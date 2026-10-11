package config

import (
	"os"
)

type Config struct {
	PORT   string
	SENTRY string
}

func Load() Config {
	port := os.Getenv("PORT")
	sentryDSN := os.Getenv("SENTRY_DSN")

	if port == "" {
		port = ":8080"
	}
	if sentryDSN != "" {
		sentryDSN = ""
	}
	return Config{PORT: ":" + port, SENTRY: sentryDSN}
}
