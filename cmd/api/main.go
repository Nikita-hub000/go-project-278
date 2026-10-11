package main

import (
	"log"
	"net/http"

	"links/internal/config"
	"links/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if err := setupRouter().Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	return router
}
