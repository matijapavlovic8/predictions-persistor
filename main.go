package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.CreateDatabase(); err != nil {
		log.Fatal(err)
	}

	server := NewServer(":8080", store)
	server.Run()

}
