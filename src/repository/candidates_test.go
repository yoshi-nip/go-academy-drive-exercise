package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestCandidateUsers_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}
	require.NoError(t, godotenv.Load("../.env"))
	db, err := ConnectDB(DBConfig{
		Host: os.Getenv("READ_DB_HOST"), Port: os.Getenv("READ_DB_PORT"),
		User: os.Getenv("READ_DB_USER"), Password: os.Getenv("READ_DB_PASSWORD"), DBName: os.Getenv("READ_DB_NAME"),
	})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	repo := NewUserRepository(db)

	t.Run("shared_series", func(t *testing.T) {
		// 複数作品を持つ実在ユーザーを選び、固定IDへの依存を避ける。
		var id uint64
		require.NoError(t, db.QueryRowContext(ctx, `SELECT user_id FROM user_watched_series
 GROUP BY user_id HAVING COUNT(*) > 1 ORDER BY user_id LIMIT 1`).Scan(&id))
		// 本体の自己JOIN・DISTINCT・LIMITとは別に、該当明細をGo側で重複排除する。
		rows, err := db.QueryContext(ctx, `SELECT user_id FROM user_watched_series
 WHERE series_id IN (SELECT series_id FROM user_watched_series WHERE user_id = ?)`, id)
		require.NoError(t, err)
		defer rows.Close()
		eligible := map[uint64]bool{}
		for rows.Next() {
			var other uint64
			require.NoError(t, rows.Scan(&other))
			if other != id {
				eligible[other] = true
			}
		}
		require.NoError(t, rows.Err())
		require.NoError(t, rows.Close())
		require.Greater(t, len(eligible), 10, "上限を検証するため共通作品の相手が11人以上いるデータが必要")
		got, err := repo.GetSharedSeriesUsers(ctx, id)
		require.NoError(t, err)
		require.Len(t, got, 10)
		seen := map[uint64]bool{}
		for _, candidate := range got {
			require.NotEqual(t, id, candidate.UserID, "自分を含めない")
			require.False(t, seen[candidate.UserID], "同じ人を重複させない")
			require.True(t, eligible[candidate.UserID], "共通作品がある相手だけ返す")
			seen[candidate.UserID] = true
		}
	})
	t.Run("empty", func(t *testing.T) {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_watched_series WHERE user_id = 0").Scan(&n))
		require.Zero(t, n)
		got, err := repo.GetSharedSeriesUsers(ctx, 0)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})
	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetSharedSeriesUsers(canceled, 1)
		require.ErrorIs(t, err, context.Canceled)
	})
}
