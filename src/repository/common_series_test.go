package repository

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
)

func TestGetCommonSeries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}
	require.NoError(t, godotenv.Load("../.env"))
	db, err := ConnectDB(DBConfig{Host: os.Getenv("READ_DB_HOST"), Port: os.Getenv("READ_DB_PORT"), User: os.Getenv("READ_DB_USER"), Password: os.Getenv("READ_DB_PASSWORD"), DBName: os.Getenv("READ_DB_NAME")})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 2人分を自己JOINする本体SQLと独立して、全明細の積集合をGo側で求める。
	rows, err := db.QueryContext(ctx, `SELECT w.user_id, s.id, s.title, s.genre_id
 FROM user_watched_series w JOIN series s ON s.id = w.series_id`)
	require.NoError(t, err)
	defer rows.Close()
	watched := map[uint64]map[string]model.WatchSeries{}
	for rows.Next() {
		var id uint64
		var s model.WatchSeries
		require.NoError(t, rows.Scan(&id, &s.SeriesID, &s.Title, &s.GenreID))
		if watched[id] == nil {
			watched[id] = map[string]model.WatchSeries{}
		}
		watched[id][s.SeriesID] = s
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	var ids []uint64
	for id := range watched {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var base uint64
	for _, id := range ids {
		if len(watched[id]) > 5 {
			base = id
			break
		}
	}
	require.NotZero(t, base, "表示上限5件を超えるユーザーが必要")
	intersection := func(other uint64) []model.WatchSeries {
		result := []model.WatchSeries{}
		for id, s := range watched[base] {
			if _, ok := watched[other][id]; ok {
				result = append(result, s)
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].SeriesID < result[j].SeriesID })
		return result
	}
	var partial, disjoint uint64
	for _, id := range ids {
		if id == base {
			continue
		}
		n := len(intersection(id))
		// どちらか一方の全作品を返す誤りを検出できる相手を選ぶ。
		if partial == 0 && n > 0 && n < len(watched[base]) && n < len(watched[id]) {
			partial = id
		}
		if disjoint == 0 && n == 0 {
			disjoint = id
		}
		if partial != 0 && disjoint != 0 {
			break
		}
	}
	require.NotZero(t, partial, "双方に非共通作品があるペアが必要")
	require.NotZero(t, disjoint, "共通作品なしの実在ユーザーが必要")
	repo := NewUserRepository(db)
	for _, tc := range []struct {
		name  string
		other uint64
	}{
		{"partial_overlap", partial}, {"no_common_series", disjoint}, {"same_user_all_series", base},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetCommonSeries(ctx, base, tc.other)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, intersection(tc.other), got, "作品・タイトル・ジャンル・順序が一致する")
		})
	}
	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetCommonSeries(canceled, base, partial)
		require.ErrorIs(t, err, context.Canceled)
	})
}
