package handler

import (
	"context"
	"go-academy-drive-exercise/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubDemoUserReader struct {
	users []model.DemoUserID
	err   error
}

func (s *stubDemoUserReader) GetDemoUsers(
	ctx context.Context,
) ([]model.DemoUserID, error) {
	return s.users, s.err
}
func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(Handlers{Users: NewUserHandler(&stubDemoUserReader{}, &stubWatchSeriesReader{}, &stubWatchGenreReader{})})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
