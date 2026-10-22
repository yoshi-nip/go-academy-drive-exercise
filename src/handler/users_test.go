package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-academy-drive-exercise/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDemoUsersHandler(t *testing.T) {

	gin.SetMode(gin.TestMode)

	reader := &stubDemoUserReader{
		users: []repository.DemoUserID{{UserID: 1}, {UserID: 2}, {UserID: 3}, {UserID: 4}, {UserID: 5}},
	}

	userHandler := NewUserHandler(reader)
	r := NewRouter(Handlers{Users: userHandler})
	// r.GET("/demo-users", GetDemoUsersHandler(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/demo-users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")

	var resp struct {
		Users []struct {
			UserID int `json:"UserID"`
		} `json:"users"`
	}

	var err error
	err = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.NoError(t, err)
	assert.Len(t, resp.Users, 5)
	for _, u := range resp.Users {
		assert.Greater(t, u.UserID, 0)
	}
}

func TestGetDemoUsersHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reader := &stubDemoUserReader{
		err: errors.New("取得失敗"),
	}
	userHandler := NewUserHandler(reader)
	r := NewRouter(Handlers{Users: userHandler})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/demo-users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t,
		`{"error":"demoユーザーの取得に失敗しました"}`,
		w.Body.String(),
	)
}
