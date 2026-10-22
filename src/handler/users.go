package handler

import (
	"context"
	"go-academy-drive-exercise/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DemoUserReader interface {
	GetDemoUsers(ctx context.Context) ([]repository.DemoUserID, error)
}

type UserHandler struct {
	reader DemoUserReader // IF
}

// NewUserHandler は UserHandler のコンストラクタで reader(IF)を受け取る
func NewUserHandler(reader DemoUserReader) *UserHandler {
	return &UserHandler{
		reader: reader,
	}
}

func (h *UserHandler) GetDemoUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		users, err := h.reader.GetDemoUsers(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "demoユーザーの取得に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}
