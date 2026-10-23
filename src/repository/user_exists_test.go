package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestUserExists_Integration(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var existingID uint64
	require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM users ORDER BY id LIMIT 1").Scan(&existingID), "投入済みユーザーが必要")
	// 正の不存在IDをデータから選ぶ。DBへの書き込みは行わない。
	var missingID uint64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT u.id + 1 FROM users u
 WHERE u.id < 18446744073709551615
 AND NOT EXISTS (SELECT 1 FROM users n WHERE n.id = u.id + 1)
 ORDER BY u.id LIMIT 1`).Scan(&missingID))
	repo := NewUserRepository(db)
	for _, tc := range []struct {
		name string
		id   uint64
		want bool
	}{
		{"exists", existingID, true}, {"missing", missingID, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.UserExists(ctx, tc.id)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
