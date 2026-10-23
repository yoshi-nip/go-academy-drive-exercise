package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-academy-drive-exercise/model"

	"github.com/stretchr/testify/require"
)

type candidateSourceStub struct {
	check  func(context.Context, uint64) (bool, error)
	get    func(context.Context, uint64) ([]model.UserID, error)
	genres func(context.Context, uint64, int) ([]model.UserID, error)
	others func(context.Context, uint64, int) ([]model.UserID, error)
}

func (s candidateSourceStub) UserExists(c context.Context, id uint64) (bool, error) {
	return s.check(c, id)
}
func (s candidateSourceStub) GetSharedSeriesUsers(c context.Context, id uint64) ([]model.UserID, error) {
	return s.get(c, id)
}
func (s candidateSourceStub) GetSharedGenreUsers(c context.Context, id uint64, limit int) ([]model.UserID, error) {
	return s.genres(c, id, limit)
}

func (s candidateSourceStub) GetOtherUsers(c context.Context, id uint64, limit int) ([]model.UserID, error) {
	return s.others(c, id, limit)
}

func TestGetCandidateUsers(t *testing.T) {
	failure := errors.New("read failed")
	ids := func(start, count int) []model.UserID {
		result := make([]model.UserID, 0, count)
		for i := 0; i < count; i++ {
			result = append(result, model.UserID{UserID: uint64(start + i)})
		}
		return result
	}
	for _, tc := range []struct {
		name                                             string
		reasons                                          string
		exists                                           bool
		series, genres, others, want                     []model.UserID
		checkErr, seriesErr, genreErr, otherErr, wantErr error
		calls                                            []string
		limit                                            int
		otherLimit                                       int
	}{
		{name: "full_skips_genres", reasons: "shared_series shared_series shared_series shared_series shared_series shared_series shared_series shared_series shared_series shared_series", exists: true, series: ids(1, 10), want: ids(1, 10), calls: []string{"check", "series"}},
		{name: "fill_remaining", reasons: "shared_series shared_series shared_series shared_series shared_series shared_series shared_series shared_genre shared_genre shared_genre", exists: true, series: ids(1, 7), genres: ids(8, 3), want: ids(1, 10), limit: 3, calls: []string{"check", "series", "genres"}},
		{name: "fill_other", reasons: "shared_series shared_series shared_series shared_series shared_series shared_series shared_series shared_genre shared_genre other", exists: true, series: ids(1, 7), genres: ids(8, 2), others: ids(10, 1), want: ids(1, 10), limit: 3, otherLimit: 1, calls: []string{"check", "series", "genres", "others"}},
		{name: "insufficient", reasons: "shared_series shared_series shared_series shared_series shared_genre shared_genre other other", exists: true, series: ids(1, 4), genres: ids(5, 2), others: ids(7, 2), want: ids(1, 8), limit: 6, otherLimit: 4, calls: []string{"check", "series", "genres", "others"}},
		{name: "only_other", reasons: "other other other other other other other other other other", exists: true, series: ids(1, 0), genres: ids(1, 0), others: ids(1, 10), want: ids(1, 10), limit: 10, otherLimit: 10, calls: []string{"check", "series", "genres", "others"}},
		{name: "other_error_discards_partial_result", exists: true, series: ids(1, 7), genres: ids(8, 2), otherErr: failure, wantErr: failure, limit: 3, otherLimit: 1, calls: []string{"check", "series", "genres", "others"}},
		{name: "empty", exists: true, series: ids(1, 0), genres: ids(1, 0), want: ids(1, 0), limit: 10, otherLimit: 10, calls: []string{"check", "series", "genres", "others"}},
		{name: "missing", wantErr: ErrUserNotFound, calls: []string{"check"}},
		{name: "check_error", checkErr: failure, wantErr: failure, calls: []string{"check"}},
		{name: "series_error", exists: true, seriesErr: failure, wantErr: failure, calls: []string{"check", "series"}},
		{name: "genre_error_discards_partial_result", exists: true, series: ids(1, 7), genreErr: failure, wantErr: failure, limit: 3, calls: []string{"check", "series", "genres"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			record := func(name string, c context.Context, id uint64) {
				require.Equal(t, ctx, c)
				require.Equal(t, uint64(42), id)
				calls = append(calls, name)
			}
			source := candidateSourceStub{
				check: func(c context.Context, id uint64) (bool, error) {
					record("check", c, id)
					return tc.exists, tc.checkErr
				},
				get: func(c context.Context, id uint64) ([]model.UserID, error) {
					record("series", c, id)
					return tc.series, tc.seriesErr
				},
				genres: func(c context.Context, id uint64, limit int) ([]model.UserID, error) {
					record("genres", c, id)
					require.Equal(t, tc.limit, limit)
					return tc.genres, tc.genreErr
				},
				others: func(c context.Context, id uint64, limit int) ([]model.UserID, error) {
					record("others", c, id)
					require.Equal(t, tc.otherLimit, limit)
					return tc.others, tc.otherErr
				},
			}
			got, err := NewCandidateService(source, source).GetCandidateUsers(ctx, 42)
			require.Equal(t, tc.calls, calls)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(tc.want))
			wantReasons := strings.Fields(tc.reasons)
			require.Len(t, wantReasons, len(tc.want))
			for i, user := range tc.want {
				require.Equal(t, user.UserID, got[i].UserID)
				require.Equal(t, wantReasons[i], got[i].Reason)
			}
		})
	}
}
