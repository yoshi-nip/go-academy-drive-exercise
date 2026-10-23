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

func TestGetCommonGenres_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}
	require.NoError(t, godotenv.Load("../.env"))
	db, err := ConnectDB(DBConfig{Host: os.Getenv("READ_DB_HOST"), Port: os.Getenv("READ_DB_PORT"), User: os.Getenv("READ_DB_USER"), Password: os.Getenv("READ_DB_PASSWORD"), DBName: os.Getenv("READ_DB_NAME")})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 本体のジャンル同士のJOINとは別に、各ユーザーの明細から期待値を作る。
	rows, err := db.QueryContext(ctx, `SELECT w.user_id, s.genre_id FROM user_watched_series w JOIN series s ON s.id = w.series_id`)
	require.NoError(t, err)
	defer rows.Close()
	genres := map[uint64]map[string]int{}
	for rows.Next() {
		var id uint64
		var genre string
		require.NoError(t, rows.Scan(&id, &genre))
		if genres[id] == nil {
			genres[id] = map[string]int{}
		}
		genres[id][genre]++
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	var ids []uint64
	for id := range genres {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	common := func(a, b uint64) []model.Genre {
		result := []model.Genre{}
		for g := range genres[a] {
			if g != "" && genres[b][g] > 0 {
				result = append(result, model.Genre{GenreID: g})
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].GenreID < result[j].GenreID })
		return result
	}
	var base, partial, emptyGenre uint64
	for _, id := range ids {
		duplicate := false
		for g, n := range genres[id] {
			if g != "" && n > 1 {
				duplicate = true
			}
		}
		if base == 0 && len(common(id, id)) >= 2 && duplicate {
			base = id
		}
		if emptyGenre == 0 && genres[id][""] > 0 {
			emptyGenre = id
		}
	}
	require.NotZero(t, base, "複数ジャンルと同ジャンルの複数作品を持つユーザーが必要")
	for _, id := range ids {
		n := len(common(base, id))
		if id != base && n > 0 && n < len(common(base, base)) {
			partial = id
			break
		}
	}
	require.NotZero(t, partial, "基準ユーザーの一部ジャンルだけを共有する相手が必要")
	repo := NewUserRepository(db)
	for _, tc := range []struct {
		name string
		a, b uint64
	}{
		{"partial_overlap", base, partial}, {"reverse_users", partial, base}, {"deduplicated_and_sorted", base, base},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetCommonGenres(ctx, tc.a, tc.b)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, common(tc.a, tc.b), got)
		})
	}
	t.Run("empty_genre_excluded", func(t *testing.T) {
		if emptyGenre == 0 {
			t.Skip("空ジャンルの視聴記録がないため実データでは検証不可")
		}
		got, err := repo.GetCommonGenres(ctx, emptyGenre, emptyGenre)
		require.NoError(t, err)
		require.Equal(t, common(emptyGenre, emptyGenre), got)
	})
	t.Run("empty", func(t *testing.T) {
		require.NotContains(t, genres, uint64(0), "0に視聴記録がないことを確認")
		got, err := repo.GetCommonGenres(ctx, base, 0)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})
	t.Run("canceled", func(t *testing.T) {
		canceled, stop := context.WithCancel(context.Background())
		stop()
		_, err := repo.GetCommonGenres(canceled, base, partial)
		require.ErrorIs(t, err, context.Canceled)
	})
}
