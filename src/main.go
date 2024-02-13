package main

import (
	// "fmt"
	"log"
)

func main() {
	storage, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", storage)
	server.Run()
}
