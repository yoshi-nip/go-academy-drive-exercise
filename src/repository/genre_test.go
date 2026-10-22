package repository

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"go-academy-drive-exercise/model"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestGetWatchGenreCounts_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}
	require.NoError(t, godotenv.Load("../.env"))
	db, err := ConnectDB(DBConfig{
		Host: os.Getenv("READ_DB_HOST"), Port: os.Getenv("READ_DB_PORT"),
		User: os.Getenv("READ_DB_USER"), Password: os.Getenv("READ_DB_PASSWORD"),
		DBName: os.Getenv("READ_DB_NAME"),
	})

	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repo := NewUserRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 表示上限5件を超えるユーザーを選び、集計SQLとは別に明細から期待値を作る。
	var userID uint64
	err = db.QueryRowContext(ctx, `SELECT user_id FROM user_watched_series
 GROUP BY user_id HAVING COUNT(*) > 5 ORDER BY user_id LIMIT 1`).Scan(&userID)
	require.NoError(t, err, "この統合テストは視聴作品が6件以上ある投入済みデータを必要とする")
	rows, err := db.QueryContext(ctx, `SELECT s.genre_id FROM user_watched_series w
 JOIN series s ON s.id = w.series_id WHERE w.user_id = ?`, userID)
	require.NoError(t, err)
	counts := map[string]uint64{}
	for rows.Next() {
		var genre string
		require.NoError(t, rows.Scan(&genre))
		if genre != "" {
			counts[genre]++
		}
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	want := make([]model.GenreWatchCount, 0, len(counts))
	for genre, count := range counts {
		want = append(want, model.GenreWatchCount{GenreID: genre, SeriesCount: count})
	}
	sort.Slice(want, func(i, j int) bool {
		if want[i].SeriesCount != want[j].SeriesCount {
			return want[i].SeriesCount > want[j].SeriesCount
		}
		return want[i].GenreID < want[j].GenreID
	})
	got, err := repo.GetWatchGenreCounts(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, want, got)

	// デバッグ表示
	t.Logf("userID=%d の視聴ジャンル集計: %+v", userID, got)

	t.Run("no_records", func(t *testing.T) {
		// 0が実データにないことを確認してから空結果を検証する。
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_watched_series WHERE user_id = 0").Scan(&n))
		require.Zero(t, n)
		got, err := repo.GetWatchGenreCounts(ctx, 0)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})

	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetWatchGenreCounts(canceled, userID)
		require.ErrorIs(t, err, context.Canceled)
	})
}
