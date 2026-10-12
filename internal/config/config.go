package config

import (
	"os"
)

type Config struct {
	PORT        string
	SENTRY      string
	DatabaseURL string
	HostURL     string
}

func Load() Config {
	port := os.Getenv("PORT")
	sentryDSN := os.Getenv("SENTRY_DSN")
	hostUrl := os.Getenv("HOST_URL")

	if port == "" {
		port = "8080"
	}
	if sentryDSN != "" {
		sentryDSN = ""
	}
	return Config{PORT: ":" + port, SENTRY: sentryDSN, DatabaseURL: os.Getenv("DATABASE_URL"), HostURL: hostUrl}
}
