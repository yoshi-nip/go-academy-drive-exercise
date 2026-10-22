package repository

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetDemoUsers_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}

	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	cfg := DBConfig{
		Host:     os.Getenv("READ_DB_HOST"),
		Port:     os.Getenv("READ_DB_PORT"),
		User:     os.Getenv("READ_DB_USER"),
		Password: os.Getenv("READ_DB_PASSWORD"),
		DBName:   os.Getenv("READ_DB_NAME"),
	}

	db, err := ConnectDB(cfg)
	if err != nil {
		t.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	repo := NewUserRepository(db)
	users, err := repo.GetDemoUsers(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Greater(t, len(users), 0, "デモユーザーが0件")
	assert.LessOrEqual(t, len(users), 5, "デモユーザーが5件を超えている")

	// デバッグ表示
	t.Logf("取得したデモユーザー: %+v", users)
}

func TestGetWatchSeries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}

	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	cfg := DBConfig{
		Host:     os.Getenv("READ_DB_HOST"),
		Port:     os.Getenv("READ_DB_PORT"),
		User:     os.Getenv("READ_DB_USER"),
		Password: os.Getenv("READ_DB_PASSWORD"),
		DBName:   os.Getenv("READ_DB_NAME"),
	}

	db, err := ConnectDB(cfg)
	if err != nil {
		t.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewUserRepository(db)

	// 実在するuserIDを指定（テストデータに合わせて調整すること
	userID := uint64(196606)
	series, err := repo.GetWatchSeries(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, series)
	assert.LessOrEqual(t, len(series), 5, "視聴シリーズが5件を超えている")

	// 取得したデータの検証
	for _, s := range series {
		assert.NotEmpty(t, s.SeriesID, "SeriesIDが空")
		assert.NotEmpty(t, s.Title, "Titleが空")
	}

	// デバッグ表示
	t.Logf("userID=%d の視聴シリーズ: %+v", userID, series)
}
