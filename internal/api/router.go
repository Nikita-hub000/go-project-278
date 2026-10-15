package api

import (
	"database/sql"
	"net/http"

	"links/internal/links"
	"links/internal/middleware"
	"links/internal/visits"

	"github.com/gin-gonic/gin"
)

func NewRouter(db *sql.DB) *gin.Engine {
	router := gin.New()
	cors := middleware.CORSMiddleware()

	router.Use(middleware.Logger(), gin.Recovery(), cors)
	router.TrustedPlatform = gin.PlatformCloudflare

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/api/links", links.GetAllLinks(db))
	router.POST("/api/links", links.PostLink(db))
	router.GET("/api/links/:id", links.GetLinkByID(db))
	router.PUT("/api/links/:id", links.UpdateLink(db))
	router.DELETE("/api/links/:id", links.DeleteLink(db))
	router.GET("/api/link_visits", visits.GetAllVisits(db))
	router.GET("/r/:code", visits.RedirectToShortLink(db))
	return router
}
