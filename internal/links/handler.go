package links

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PostLink(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := CreateLinkRequest{}
		if !bindLinkRequest(c, &request) {
			return
		}

		link, err := CreateShortLink(c.Request.Context(), db, request)
		if err != nil {
			writeLinkError(c, err, "failed to create link")
			return
		}

		c.JSON(http.StatusCreated, link)
	}
}

func GetAllLinks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := c.Query("range")
		filter := c.Query("filter")
		sort := c.Query("sort")
		links, err := GetAllLinksHandler(c.Request.Context(), db, pagination, sort, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve links"})
			return
		}
		if pagination != "" {
			c.Header("Content-Range", fmt.Sprintf("links %s/%d", pagination, len(links)))
		}
		if links == nil {
			links = []Link{}
		}
		c.JSON(http.StatusOK, links)
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

		c.JSON(http.StatusOK, link)
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
		if !bindLinkRequest(c, &request) {
			return
		}

		link, err := UpdateShortLink(c.Request.Context(), db, id, request)

		if err != nil {
			writeLinkError(c, err, "failed to update link")
			return
		}

		c.JSON(http.StatusOK, link)
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
