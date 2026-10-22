package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-academy-drive-exercise/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetDemoUsersHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}

	gin.SetMode(gin.TestMode)

	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("ローディングエラー .env file: %v", err)
	}

	cfg := repository.DBConfig{
		Host:     os.Getenv("READ_DB_HOST"),
		Port:     os.Getenv("READ_DB_PORT"),
		User:     os.Getenv("READ_DB_USER"),
		Password: os.Getenv("READ_DB_PASSWORD"),
		DBName:   os.Getenv("READ_DB_NAME"),
	}

	db, err := repository.ConnectDB(cfg)
	if err != nil {
		t.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	r := NewRouter(db)
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

	err = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.NoError(t, err)
	assert.Len(t, resp.Users, 5)
	for _, u := range resp.Users {
		assert.Greater(t, u.UserID, 0)
	}
}
