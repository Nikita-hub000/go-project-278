package visits

import (
	"database/sql"
	"errors"
	"links/internal/pagination"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToShortLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		url, err := CreateVisit(c.Request.Context(), db, code, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			if url == "" {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
					return
				}
				slog.Error("redirect lookup failed", "operation", "find_link", "code", code, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create visit"})
				return
			}
			slog.Error("visit recording failed", "operation", "create_visit", "code", code, "error", err)
		}
		c.Header("Cache-Control", "no-store")
		c.Redirect(http.StatusFound, url)
	}
}

func GetAllVisits(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rangeQuery := c.Query("range")
		sort := c.Query("sort")
		filter := c.Query("filter")
		visits, total, err := GetAllVisitsHandler(c.Request.Context(), db, rangeQuery, sort, filter)
		if err != nil {
			if errors.Is(err, pagination.ErrInvalidRange) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve visits"})
			return
		}
		var offset int32
		if rangeQuery != "" {
			page, _ := pagination.Parse(rangeQuery)
			offset = page.Offset
		}
		c.Header("Content-Range", pagination.ContentRange("link_visits", offset, len(visits), total))
		c.JSON(http.StatusOK, visits)
	}
}
