package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"links/internal/config"
	"links/internal/middleware"
	"links/internal/platform/postgres"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("database initialization: %v", err)
	}
	defer db.Close()

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
	if err := setupRouter().Run(cfg.PORT); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		//sentry.CaptureMessage("Error created on purpose to send it to Bugsink")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	return router
}
