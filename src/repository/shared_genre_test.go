package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestGetSharedGenreUsers_Integration(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repo := NewUserRepository(db)

	// SQLの自己JOINやNOT EXISTSを写さず、明細をGo側で比較して期待する候補集合を作る。
	type watch struct {
		user          uint64
		series, genre string
	}
	rows, err := db.QueryContext(ctx, `SELECT w.user_id, w.series_id, s.genre_id
 FROM user_watched_series w JOIN series s ON s.id = w.series_id`)
	require.NoError(t, err)
	defer rows.Close()
	var watches []watch
	var userID uint64
	for rows.Next() {
		var w watch
		require.NoError(t, rows.Scan(&w.user, &w.series, &w.genre))
		watches = append(watches, w)
		if w.genre != "" && (userID == 0 || w.user < userID) {
			userID = w.user
		}
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.NotZero(t, userID, "空でないジャンルの視聴記録を持つユーザーが必要")
	mySeries, myGenres := map[string]bool{}, map[string]bool{}
	for _, w := range watches {
		if w.user == userID {
			mySeries[w.series] = true
			if w.genre != "" {
				myGenres[w.genre] = true
			}
		}
	}
	sharedSeries, sharedGenres := map[uint64]bool{}, map[uint64]bool{}
	for _, w := range watches {
		if mySeries[w.series] {
			sharedSeries[w.user] = true
		}
		if w.genre != "" && myGenres[w.genre] {
			sharedGenres[w.user] = true
		}
	}
	eligible := map[uint64]bool{}
	for id := range sharedGenres {
		if id != userID && !sharedSeries[id] {
			eligible[id] = true
		}
	}
	require.Greater(t, len(eligible), 3, "人数制限を検証するためジャンルのみ共通の相手が4人以上いるデータが必要")
	t.Logf("基準userID=%d、ジャンルのみ共通の候補=%d人", userID, len(eligible))

	for _, limit := range []int{1, 3} {
		t.Run(fmt.Sprintf("limit_%d", limit), func(t *testing.T) {
			got, err := repo.GetSharedGenreUsers(ctx, userID, limit)
			require.NoError(t, err)
			require.Len(t, got, limit)
			seen := map[uint64]bool{}
			for _, candidate := range got {
				id := candidate.UserID
				require.NotEqual(t, userID, id)
				require.False(t, seen[id], "候補が重複している")
				require.True(t, eligible[id], "空でない共通ジャンルがあり、共通作品がない相手だけ返す")
				seen[id] = true
			}
		})
	}
	t.Run("empty", func(t *testing.T) {
		var n int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_watched_series WHERE user_id = 0").Scan(&n))
		require.Zero(t, n)
		got, err := repo.GetSharedGenreUsers(ctx, 0, 3)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})
	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetSharedGenreUsers(canceled, userID, 3)
		require.ErrorIs(t, err, context.Canceled)
	})
}
