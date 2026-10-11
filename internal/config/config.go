package config

import "os"

// Config contains application settings loaded from the environment.
type Config struct {
	HTTPAddr string
}

// Load returns environment settings with defaults for local development.
func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{HTTPAddr: addr}
}
