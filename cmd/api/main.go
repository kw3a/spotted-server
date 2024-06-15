package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kw3a/spotted-server/internal/server"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	log.Fatal(server.Run())
}
