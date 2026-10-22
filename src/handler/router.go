package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	r.GET("/health", HealthHandler)
	r.GET("/demo-users", GetDemoUsersHandler(db))
	return r
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
