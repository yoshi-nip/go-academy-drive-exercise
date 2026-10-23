package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
	"go-academy-drive-exercise/service"
)

type stubViewingProfile struct {
	get func(context.Context, uint64) (model.ViewingProfile, error)
}

func (s stubViewingProfile) GetViewingProfile(ctx context.Context, id uint64) (model.ViewingProfile, error) {
	return s.get(ctx, id)
}

func TestGetViewingProfileHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, id      string
		profile       model.ViewingProfile
		err           error
		status, calls int
		body          string
	}{
		{name: "success", id: "42", profile: model.ViewingProfile{UserID: 42, Series: []model.WatchSeries{{SeriesID: "s1", Title: "作品", GenreID: "g1"}}, Genres: []model.GenreWatchCount{{GenreID: "g1", SeriesCount: 6}}}, status: 200, calls: 1, body: `{"user_id":42,"series":[{"series_id":"s1","title":"作品","genre_id":"g1"}],"genres":[{"genre_id":"g1","series_count":6}]}`},
		{name: "empty", id: "42", profile: model.ViewingProfile{UserID: 42, Series: []model.WatchSeries{}, Genres: []model.GenreWatchCount{}}, status: 200, calls: 1, body: `{"user_id":42,"series":[],"genres":[]}`},
		{name: "not_found", id: "42", err: fmt.Errorf("lookup: %w", service.ErrUserNotFound), status: 404, calls: 1, body: `{"error":"ユーザーが存在しません"}`},
		{name: "error", id: "42", err: errors.New("internal detail"), status: 500, calls: 1, body: `{"error":"視聴プロフィールの取得に失敗しました"}`},
		{name: "letters", id: "abc", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "zero", id: "0", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "negative", id: "-1", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "overflow", id: "18446744073709551616", status: 400, body: `{"error":"userIDが不正です"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			source := stubViewingProfile{get: func(gotCtx context.Context, id uint64) (model.ViewingProfile, error) {
				calls++
				require.Equal(t, uint64(42), id)
				require.Equal(t, ctx, gotCtx)
				return tc.profile, tc.err
			}}
			router := NewRouter(Handlers{Users: NewUserHandler(nil, nil, nil, source, nil)})
			req := httptest.NewRequest(http.MethodGet, "/users/"+tc.id+"/viewing-profile", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			require.JSONEq(t, tc.body, w.Body.String())
			require.Equal(t, tc.calls, calls)
		})
	}
}
