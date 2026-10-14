package visits

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToShortLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		url, err := CreateVisit(c.Request.Context(), db, code, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create visit"})
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Redirect(http.StatusFound, url)
	}
}

func GetAllVisits(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := c.Query("range")
		sort := c.Query("sort")
		filter := c.Query("filter")
		visits, err := GetAllVisitsHandler(c.Request.Context(), db, pagination, sort, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve visits"})
			return
		}
		if pagination != "" {
			c.Header("Content-Range", fmt.Sprintf("visits %s/%d", pagination, len(visits)))
		}
		c.JSON(http.StatusOK, visits)
	}
}
