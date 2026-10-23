package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go-academy-drive-exercise/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDemoUsersHandler(t *testing.T) {

	gin.SetMode(gin.TestMode)

	reader := &stubDemoUserReader{
		users: []model.DemoUserID{{UserID: 1}, {UserID: 2}, {UserID: 3}, {UserID: 4}, {UserID: 5}},
	}

	userHandler := NewUserHandler(reader, &stubWatchSeriesReader{}, &stubWatchGenreReader{}, nil)
	r := NewRouter(Handlers{Users: userHandler})
	// r.GET("/demo-users", GetDemoUsersHandler(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/demo-users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")

	var resp struct {
		Users []struct {
			UserID uint64 `json:"user_id"`
		} `json:"users"`
	}

	var err error
	err = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.NoError(t, err)
	assert.Len(t, resp.Users, 5)
	for _, u := range resp.Users {
		assert.Greater(t, u.UserID, uint64(0))
	}
}

func TestGetDemoUsersHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reader := &stubDemoUserReader{
		err: errors.New("取得失敗"),
	}
	userHandler := NewUserHandler(reader, &stubWatchSeriesReader{}, &stubWatchGenreReader{}, nil)
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

type stubWatchSeriesReader struct {
	series []model.WatchSeries
	err    error
	calls  int
	userID uint64
}

func (s *stubWatchSeriesReader) GetWatchSeries(ctx context.Context, userID uint64) ([]model.WatchSeries, error) {
	s.calls++
	s.userID = userID
	return s.series, s.err
}

func TestGetWatchSeriesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		id     string
		wantID uint64
		series []model.WatchSeries
		err    error
		status int
		body   string
	}{
		{name: "success", id: "196606", wantID: 196606,
			series: []model.WatchSeries{{SeriesID: "001-A", Title: "作品", GenreID: ""}},
			status: http.StatusOK, body: `{"series":[{"series_id":"001-A","title":"作品","genre_id":""}]}`},
		{name: "empty", id: "1", wantID: 1, series: []model.WatchSeries{}, status: http.StatusOK, body: `{"series":[]}`},
		{name: "uint64_max", id: "18446744073709551615", wantID: ^uint64(0), series: []model.WatchSeries{}, status: http.StatusOK, body: `{"series":[]}`},
		{name: "reader_error", id: "1", wantID: 1, err: errors.New("internal DB detail"), status: http.StatusInternalServerError, body: `{"error":"視聴シリーズの取得に失敗しました"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &stubWatchSeriesReader{series: tc.series, err: tc.err}
			r := NewRouter(Handlers{Users: NewUserHandler(&stubDemoUserReader{}, reader, &stubWatchGenreReader{}, nil)})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/"+tc.id+"/watch-series", nil))
			assert.Equal(t, tc.status, w.Code)
			assert.JSONEq(t, tc.body, w.Body.String())
			assert.Equal(t, 1, reader.calls)
			assert.Equal(t, tc.wantID, reader.userID)
		})
	}
}

func TestGetWatchSeriesHandler_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, id := range []string{"abc", "-1", "0", "18446744073709551616"} {
		t.Run(id, func(t *testing.T) {
			reader := &stubWatchSeriesReader{}
			r := NewRouter(Handlers{Users: NewUserHandler(&stubDemoUserReader{}, reader, &stubWatchGenreReader{}, nil)})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/"+id+"/watch-series", nil))
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, `{"error":"userIDが不正です"}`, w.Body.String())
			assert.Zero(t, reader.calls, "不正IDでは取得処理を呼ばない")
		})
	}
}
