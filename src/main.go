package main

import "go-academy-drive-exercise/handler"

func main() {
	if err := handler.NewRouter().Run(":8080"); err != nil {
		panic(err)
	}
}
