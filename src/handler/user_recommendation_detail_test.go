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

type detailServiceStub struct {
	get func(context.Context, uint64, uint64) (*model.UserRecommendationDetail, error)
}

func (s detailServiceStub) GetUserRecommendationDetail(c context.Context, a, b uint64) (*model.UserRecommendationDetail, error) {
	return s.get(c, a, b)
}

func TestGetUserRecommendationDetailHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	detail := &model.UserRecommendationDetail{UserID: 42, OtherUserID: 99, CommonSeries: []model.WatchSeries{{SeriesID: "s1", Title: "作品", GenreID: "movie"}}, CommonGenres: []model.Genre{{GenreID: "movie"}}, RecommendedSeries: []model.WatchSeries{{SeriesID: "s2", Title: "推薦", GenreID: "animation"}}}
	empty := &model.UserRecommendationDetail{UserID: 42, OtherUserID: 99, CommonSeries: []model.WatchSeries{}, CommonGenres: []model.Genre{}, RecommendedSeries: []model.WatchSeries{}}
	for _, tc := range []struct {
		name   string
		detail *model.UserRecommendationDetail
		err    error
		status int
		body   string
	}{
		{name: "success", detail: detail, status: 200, body: `{"user_id":42,"other_user_id":99,"common_series":[{"series_id":"s1","title":"作品","genre_id":"movie"}],"common_genres":[{"genre_id":"movie"}],"recommended_series":[{"series_id":"s2","title":"推薦","genre_id":"animation"}]}`},
		{name: "empty", detail: empty, status: 200, body: `{"user_id":42,"other_user_id":99,"common_series":[],"common_genres":[],"recommended_series":[]}`},
		{name: "missing", err: fmt.Errorf("lookup: %w", service.ErrUserNotFound), status: 404, body: `{"error":"ユーザーが存在しません"}`},
		{name: "error", err: errors.New("private detail"), status: 500, body: `{"error":"ユーザー推薦詳細の取得に失敗しました"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			source := detailServiceStub{get: func(c context.Context, a, b uint64) (*model.UserRecommendationDetail, error) {
				calls++
				require.Equal(t, ctx, c)
				require.Equal(t, uint64(42), a)
				require.Equal(t, uint64(99), b)
				return tc.detail, tc.err
			}}
			router := NewRouter(Handlers{Users: NewUserHandler(nil, nil, nil, nil, nil, source)})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/42/user-recommendation-detail/99", nil).WithContext(ctx))
			require.Equal(t, tc.status, w.Code)
			require.JSONEq(t, tc.body, w.Body.String())
			require.Equal(t, 1, calls)
		})
	}
	for _, bad := range []string{"abc", "0", "-1", "18446744073709551616"} {
		for _, side := range []string{"user", "other"} {
			t.Run(side+"_"+bad, func(t *testing.T) {
				a, b := "42", "99"
				if side == "user" {
					a = bad
				} else {
					b = bad
				}
				calls := 0
				source := detailServiceStub{get: func(context.Context, uint64, uint64) (*model.UserRecommendationDetail, error) {
					calls++
					return nil, nil
				}}
				router := NewRouter(Handlers{Users: NewUserHandler(nil, nil, nil, nil, nil, source)})
				w := httptest.NewRecorder()
				router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/"+a+"/user-recommendation-detail/"+b, nil))
				require.Equal(t, 400, w.Code)
				require.JSONEq(t, `{"error":"userIDまたはotherUserIDが不正です"}`, w.Body.String())
				require.Zero(t, calls)
			})
		}
	}
}

func TestGetUserRecommendationDetailHandler_SameUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 実serviceにつなぎ、同一IDをDBアクセスなしで400へ変換できることを確認する。
	detailService := service.NewUserRecommendationDetailService(nil, nil)
	router := NewRouter(Handlers{Users: NewUserHandler(nil, nil, nil, nil, nil, detailService)})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/42/user-recommendation-detail/42", nil))
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.JSONEq(t, `{"error":"同じユーザーは指定できません"}`, w.Body.String())
}
