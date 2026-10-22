package handler

import (
	"database/sql"
	"go-academy-drive-exercise/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDemoUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetDemoUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "demoユーザーの取得に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}
