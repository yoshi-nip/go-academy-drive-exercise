package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
)

func TestGetRecommendedSeries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}
	require.NoError(t, godotenv.Load("../.env"))
	db, err := ConnectDB(DBConfig{Host: os.Getenv("READ_DB_HOST"), Port: os.Getenv("READ_DB_PORT"), User: os.Getenv("READ_DB_USER"), Password: os.Getenv("READ_DB_PASSWORD"), DBName: os.Getenv("READ_DB_NAME")})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	repo := NewUserRepository(db)
	for _, tc := range []struct{ name, condition string }{
		{"over_limit", "COUNT(*) > 5"}, {"under_limit", "COUNT(*) BETWEEN 1 AND 4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var id uint64
			// conditionはテスト内の固定値のみ。
			require.NoError(t, db.QueryRowContext(ctx, "SELECT user_id FROM user_recommended_series GROUP BY user_id HAVING "+tc.condition+" ORDER BY user_id LIMIT 1").Scan(&id))
			rows, err := db.QueryContext(ctx, "SELECT series_id FROM user_watched_series WHERE user_id = ?", id)
			require.NoError(t, err)
			defer rows.Close()
			watched := map[string]bool{}
			for rows.Next() {
				var seriesID string
				require.NoError(t, rows.Scan(&seriesID))
				watched[seriesID] = true
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			// NOT EXISTS・RAND・LIMITを使わず、明細からGo側で期待集合を作る。
			rows, err = db.QueryContext(ctx, `SELECT s.id,s.title,s.genre_id FROM user_recommended_series r JOIN series s ON s.id=r.series_id WHERE r.user_id=?`, id)
			require.NoError(t, err)
			defer rows.Close()
			eligible := map[string]model.WatchSeries{}
			for rows.Next() {
				var s model.WatchSeries
				require.NoError(t, rows.Scan(&s.SeriesID, &s.Title, &s.GenreID))
				if !watched[s.SeriesID] {
					eligible[s.SeriesID] = s
				}
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			if tc.name == "over_limit" {
				require.Greater(t, len(eligible), 5)
			} else {
				require.NotEmpty(t, eligible)
			}
			got, err := repo.GetRecommendedSeries(ctx, id)
			require.NoError(t, err)
			wantCount := len(eligible)
			if wantCount > 5 {
				wantCount = 5
			}
			require.Len(t, got, wantCount)
			seen := map[string]bool{}
			for _, s := range got {
				require.False(t, seen[s.SeriesID], "作品の重複")
				require.False(t, watched[s.SeriesID], "本人の視聴済み作品を除外")
				expected, ok := eligible[s.SeriesID]
				require.True(t, ok, "指定ユーザーへの推薦に含まれる")
				require.Equal(t, expected, s)
				seen[s.SeriesID] = true
			}
		})
	}
	t.Run("empty", func(t *testing.T) {
		var id uint64
		require.NoError(t, db.QueryRowContext(ctx, `SELECT u.id FROM users u WHERE NOT EXISTS (SELECT 1 FROM user_recommended_series r WHERE r.user_id=u.id) ORDER BY u.id LIMIT 1`).Scan(&id))
		got, err := repo.GetRecommendedSeries(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})
	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetRecommendedSeries(canceled, 1)
		require.ErrorIs(t, err, context.Canceled)
	})
}
