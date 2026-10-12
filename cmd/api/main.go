package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"links/internal/config"
	"links/internal/links"
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
	if err := setupRouter(db).Run(cfg.PORT); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

func setupRouter(db *sql.DB) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.GET("/api/links", links.GetAllLinks(db))
	router.POST("/api/links", links.PostLink(db))
	router.GET("/api/links/:id", links.GetLinkByID(db))
	router.PUT("/api/links/:id", links.UpdateLink(db))
	router.DELETE("/api/links/:id", links.DeleteLink(db))
	return router
}
