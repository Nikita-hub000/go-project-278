package main

import (
	"context"
	"log"
	"time"

	"links/internal/api"
	"links/internal/config"
	"links/internal/platform/postgres"

	"github.com/getsentry/sentry-go"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("database initialization: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	err = sentry.Init(sentry.ClientOptions{
		Dsn:                  cfg.SENTRY,
		DataCollection:       &sentry.DataCollection{},
		EnableTracing:        false,
		TracesSampleRate:     0,
		DisableClientReports: true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)
	if err := api.NewRouter(db).Run(cfg.PORT); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
