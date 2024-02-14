package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	storage, err := NewPostgresStorage()
	defer storage.Close()
	if err != nil {
		log.Fatal(err)
	}

	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	portAddress := os.Getenv("PORT")

	server := NewAPIServer(fmt.Sprintf(":%s", portAddress), storage)
	server.Run()
}
