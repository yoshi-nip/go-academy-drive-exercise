package handler

import (
	"context"
	"errors"
	"go-academy-drive-exercise/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubWatchGenreReader struct {
	genres []model.GenreWatchCount
	err    error
	calls  int
	id     uint64
}

func (s *stubWatchGenreReader) GetWatchGenreCounts(ctx context.Context, id uint64) ([]model.GenreWatchCount, error) {
	s.calls++
	s.id = id
	return s.genres, s.err
}
func TestGetWatchGenreCountsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, id      string
		genres        []model.GenreWatchCount
		err           error
		status, calls int
		body          string
	}{
		{name: "success", id: "42", genres: []model.GenreWatchCount{{GenreID: "001", SeriesCount: 8}}, status: 200, calls: 1, body: `{"genre":[{"genre_id":"001","series_count":8}]}`},
		{name: "empty", id: "42", genres: []model.GenreWatchCount{}, status: 200, calls: 1, body: `{"genre":[]}`},
		{name: "error", id: "42", genres: []model.GenreWatchCount{{GenreID: "partial", SeriesCount: 1}}, err: errors.New("internal detail"), status: 500, calls: 1, body: `{"error":"視聴ジャンル集計の取得に失敗しました"}`},
		{name: "letters", id: "abc", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "zero", id: "0", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "negative", id: "-1", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "overflow", id: "18446744073709551616", status: 400, body: `{"error":"userIDが不正です"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &stubWatchGenreReader{genres: tc.genres, err: tc.err}
			r := NewRouter(Handlers{Users: NewUserHandler(&stubDemoUserReader{}, &stubWatchSeriesReader{}, reader, nil, nil, nil)})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/"+tc.id+"/watch-genre-counts", nil))
			assert.Equal(t, tc.status, w.Code)
			assert.JSONEq(t, tc.body, w.Body.String())
			assert.Equal(t, tc.calls, reader.calls)
			if tc.calls > 0 {
				assert.Equal(t, uint64(42), reader.id)
			}
		})
	}
}
