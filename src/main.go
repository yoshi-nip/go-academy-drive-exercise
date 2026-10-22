package main

import (
	"go-academy-drive-exercise/handler"
	"go-academy-drive-exercise/repository"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfgR := repository.DBConfig{
		Host:     os.Getenv("DB_READ_HOST"),
		Port:     os.Getenv("DB_READ_PORT"),
		User:     os.Getenv("DB_READ_USER"),
		Password: os.Getenv("DB_READ_PASSWORD"),
		DBName:   os.Getenv("DB_READ_NAME"),
	}

	// cfgW := repository.DBConfig{
	// 	Host:     os.Getenv("DB_WRITE_HOST"),
	// 	Port:     os.Getenv("DB_WRITE_PORT"),
	// 	User:     os.Getenv("DB_WRITE_USER"),
	// 	Password: os.Getenv("DB_WRITE_PASSWORD"),
	// 	DBName:   os.Getenv("DB_WRITE_NAME"),
	// }

	db, err := repository.ConnectDB(cfgR)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	log.Println("DB接続成功")

	if err := handler.NewRouter().Run(":8080"); err != nil {
		panic(err)
	}

}
