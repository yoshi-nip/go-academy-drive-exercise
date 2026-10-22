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
		Host:     os.Getenv("READ_DB_HOST"),
		Port:     os.Getenv("READ_DB_PORT"),
		User:     os.Getenv("READ_DB_USER"),
		Password: os.Getenv("READ_DB_PASSWORD"),
		DBName:   os.Getenv("READ_DB_NAME"),
	}

	// cfgW := repository.DBConfig{
	// 	Host:     os.Getenv("WRITE_DB_HOST"),
	// 	Port:     os.Getenv("WRITE_DB_PORT"),
	// 	User:     os.Getenv("WRITE_DB_USER"),
	// 	Password: os.Getenv("WRITE_DB_PASSWORD"),
	// 	DBName:   os.Getenv("WRITE_DB_NAME"),
	// }

	db, err := repository.ConnectDB(cfgR)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	log.Println("DB接続成功")

	userRepo := repository.NewUserRepository(db)
	userHandler := handler.NewUserHandler(userRepo, userRepo, userRepo)
	router := handler.NewRouter(handler.Handlers{Users: userHandler})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

}
