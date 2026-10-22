package repository

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
)

func TestConnectDB_EmptyConfig(t *testing.T) {
	cfg := DBConfig{}

	db, err := ConnectDB(cfg)

	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestConnectDB_InvalidPort(t *testing.T) {
	cfg := DBConfig{
		Host:     "127.0.0.1",
		Port:     "99999", // 不正なポート
		User:     "app_reader",
		Password: "password",
		DBName:   "app_db",
	}

	db, err := ConnectDB(cfg)

	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestConnectDB_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("実DB接続テストをスキップ")
	}

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfgR := DBConfig{
		Host:     os.Getenv("READ_DB_HOST"),
		Port:     os.Getenv("READ_DB_PORT"),
		User:     os.Getenv("READ_DB_USER"),
		Password: os.Getenv("READ_DB_PASSWORD"),
		DBName:   os.Getenv("READ_DB_NAME"),
	}

	db, err := ConnectDB(cfgR)

	assert.NoError(t, err)
	assert.NotNil(t, db)

	if db != nil {
		err = db.Ping()
		assert.NoError(t, err)
	}
	defer db.Close()
}
