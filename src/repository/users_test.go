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

	users, err := GetDemoUsers(ctx, db)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Greater(t, len(users), 0, "デモユーザーが0件")
	assert.LessOrEqual(t, len(users), 5, "デモユーザーが5件を超えている")

	// デバッグ表示
	t.Logf("取得したデモユーザー: %+v", users)
}
