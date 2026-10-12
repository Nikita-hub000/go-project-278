package links

import (
	"database/sql"
	"fmt"
	"links/internal/platform/postgres/sqlc"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PostLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := CreateLinkRequest{}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data, err := GenerateShortUrl(c.Request.Context(), db, request.ShortName)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		link, err := CreateLink(c.Request.Context(),
			db,
			&Link{ShortName: data.ShortName, OriginalURL: request.OriginalURL, ShortURL: data.ShortURL})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"link": link})
	}
}

func GetAllLinks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := c.Query("range")
		links, err := GetAllLinksHandler(c.Request.Context(), db, pagination)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve links"})
			return
		}
		if pagination != "" {
			c.Header("Content-Range", fmt.Sprintf("links %s/%d", pagination, len(links)))
		}
		c.JSON(http.StatusOK, gin.H{"links": links})
	}
}

func GetLinkByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		link, err := GetLinkByIDHandler(c.Request.Context(), db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"link": link})
	}
}

func UpdateLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		request := UpdateLinkRequest{}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		link, err := UpdateLinkHandler(c.Request.Context(), db, sqlc.UpdateLinkParams{
			OriginalUrl: request.OriginalURL,
			ShortName:   request.ShortName,
			ID:          id,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"link": link})
	}
}

func DeleteLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = DeleteLinkByIDHandler(c.Request.Context(), db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})

	}
}
