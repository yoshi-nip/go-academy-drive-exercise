package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
)

type detailSourceStub struct {
	check       func(context.Context, uint64) (bool, error)
	series      func(context.Context, uint64, uint64) ([]model.WatchSeries, error)
	genres      func(context.Context, uint64, uint64) ([]model.Genre, error)
	recommended func(context.Context, uint64) ([]model.WatchSeries, error)
}

func (s detailSourceStub) UserExists(c context.Context, id uint64) (bool, error) {
	return s.check(c, id)
}
func (s detailSourceStub) GetCommonSeries(c context.Context, a, b uint64) ([]model.WatchSeries, error) {
	return s.series(c, a, b)
}
func (s detailSourceStub) GetCommonGenres(c context.Context, a, b uint64) ([]model.Genre, error) {
	return s.genres(c, a, b)
}
func (s detailSourceStub) GetRecommendedSeries(c context.Context, id uint64) ([]model.WatchSeries, error) {
	return s.recommended(c, id)
}

func TestGetUserRecommendationDetail(t *testing.T) {
	failure := errors.New("read failed")
	for _, tc := range []struct {
		name, failAt string
		missing      uint64
		empty        bool
		wantErr      error
		calls        []string
	}{
		{name: "success", calls: []string{"user", "other", "series", "genres", "recommended"}},
		{name: "empty", empty: true, calls: []string{"user", "other", "series", "genres", "recommended"}},
		{name: "user_missing", missing: 42, wantErr: ErrUserNotFound, calls: []string{"user"}},
		{name: "other_missing", missing: 99, wantErr: ErrUserNotFound, calls: []string{"user", "other"}},
		{name: "user_check_error", failAt: "user", wantErr: failure, calls: []string{"user"}},
		{name: "other_check_error", failAt: "other", wantErr: failure, calls: []string{"user", "other"}},
		{name: "series_error", failAt: "series", wantErr: failure, calls: []string{"user", "other", "series"}},
		{name: "genres_error", failAt: "genres", wantErr: failure, calls: []string{"user", "other", "series", "genres"}},
		{name: "recommended_error", failAt: "recommended", wantErr: failure, calls: []string{"user", "other", "series", "genres", "recommended"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			record := func(name string, c context.Context) error {
				require.Equal(t, ctx, c)
				calls = append(calls, name)
				if tc.failAt == name {
					return failure
				}
				return nil
			}
			series := []model.WatchSeries{{SeriesID: "common", Title: "共通作品", GenreID: "movie"}}
			genres := []model.Genre{{GenreID: "movie"}}
			recommended := []model.WatchSeries{{SeriesID: "recommendation", Title: "推薦作品", GenreID: "animation"}}
			if tc.empty {
				series = []model.WatchSeries{}
				genres = []model.Genre{}
				recommended = []model.WatchSeries{}
			}
			source := detailSourceStub{
				check: func(c context.Context, id uint64) (bool, error) {
					require.Contains(t, []uint64{42, 99}, id)
					name := "user"
					if id == 99 {
						name = "other"
					}
					return id != tc.missing, record(name, c)
				},
				series: func(c context.Context, a, b uint64) ([]model.WatchSeries, error) {
					require.Equal(t, uint64(42), a)
					require.Equal(t, uint64(99), b)
					return series, record("series", c)
				},
				genres: func(c context.Context, a, b uint64) ([]model.Genre, error) {
					require.Equal(t, uint64(42), a)
					require.Equal(t, uint64(99), b)
					return genres, record("genres", c)
				},
				recommended: func(c context.Context, id uint64) ([]model.WatchSeries, error) {
					require.Equal(t, uint64(99), id, "相手への推薦を取得する")
					return recommended, record("recommended", c)
				},
			}
			got, err := NewUserRecommendationDetailService(source, source).GetUserRecommendationDetail(ctx, 42, 99)
			require.Equal(t, tc.calls, calls)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got, "部分結果を返さない")
				return
			}
			require.NoError(t, err)
			require.Equal(t, &model.UserRecommendationDetail{UserID: 42, OtherUserID: 99, CommonSeries: series, CommonGenres: genres, RecommendedSeries: recommended}, got)
		})
	}
}

func TestGetUserRecommendationDetail_SameUser(t *testing.T) {
	// 依存を渡さなくても拒否できることにより、存在確認・DB取得の前に止まることを確認する。
	s := NewUserRecommendationDetailService(nil, nil)
	got, err := s.GetUserRecommendationDetail(context.Background(), 42, 42)
	require.ErrorIs(t, err, ErrSameUser)
	require.Nil(t, got)
}
