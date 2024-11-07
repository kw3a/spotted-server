package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kw3a/spotted-server/internal/server"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	if env == "development" {
		log.Println("Loading .env file")
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading.env file")
		}
	}
	log.Fatal(server.Run())
}
