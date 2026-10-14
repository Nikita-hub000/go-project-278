package integration

import (
	"database/sql"
	"links/internal/api"

	"github.com/gin-gonic/gin"
)

func setupRouter(db *sql.DB) *gin.Engine {
	return api.NewRouter(db)
}
