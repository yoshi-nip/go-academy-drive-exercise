package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
)

type profileSourceStub struct {
	check  func(context.Context, uint64) (bool, error)
	series func(context.Context, uint64) ([]model.WatchSeries, error)
	genres func(context.Context, uint64) ([]model.GenreWatchCount, error)
}

func (s profileSourceStub) UserExists(ctx context.Context, id uint64) (bool, error) {
	return s.check(ctx, id)
}
func (s profileSourceStub) GetWatchSeries(ctx context.Context, id uint64) ([]model.WatchSeries, error) {
	return s.series(ctx, id)
}
func (s profileSourceStub) GetWatchGenreCounts(ctx context.Context, id uint64) ([]model.GenreWatchCount, error) {
	return s.genres(ctx, id)
}

func TestGetViewingProfile(t *testing.T) {
	failure := errors.New("read failed")
	for _, tc := range []struct {
		name      string
		exists    bool
		failAt    string
		empty     bool
		wantErr   error
		wantCalls []string
	}{
		{name: "success", exists: true, wantCalls: []string{"exists", "series", "genres"}},
		{name: "empty", exists: true, empty: true, wantCalls: []string{"exists", "series", "genres"}},
		{name: "not_found", wantErr: ErrUserNotFound, wantCalls: []string{"exists"}},
		{name: "exists_error", failAt: "exists", wantErr: failure, wantCalls: []string{"exists"}},
		{name: "series_error", exists: true, failAt: "series", wantErr: failure, wantCalls: []string{"exists", "series"}},
		{name: "genres_error", exists: true, failAt: "genres", wantErr: failure, wantCalls: []string{"exists", "series", "genres"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			const userID uint64 = 42
			calls := []string{}
			record := func(name string, gotCtx context.Context, gotID uint64) error {
				require.Equal(t, ctx, gotCtx)
				require.Equal(t, userID, gotID)
				calls = append(calls, name)
				if tc.failAt == name {
					return failure
				}
				return nil
			}
			series := []model.WatchSeries{{SeriesID: "s1", Title: "作品", GenreID: "animation"}}
			genres := []model.GenreWatchCount{{GenreID: "animation", SeriesCount: 6}}
			if tc.empty {
				series = []model.WatchSeries{}
				genres = []model.GenreWatchCount{}
			}
			source := profileSourceStub{
				check: func(c context.Context, id uint64) (bool, error) { return tc.exists, record("exists", c, id) },
				series: func(c context.Context, id uint64) ([]model.WatchSeries, error) {
					return series, record("series", c, id)
				},
				genres: func(c context.Context, id uint64) ([]model.GenreWatchCount, error) {
					return genres, record("genres", c, id)
				},
			}
			s := NewViewingProfileService(source, source, source)
			got, err := s.GetViewingProfile(ctx, userID)
			require.Equal(t, tc.wantCalls, calls)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Equal(t, model.ViewingProfile{}, got, "失敗時に部分結果を返さない")
				return
			}
			require.NoError(t, err)
			require.Equal(t, model.ViewingProfile{UserID: userID, Series: series, Genres: genres}, got)
		})
	}
}
