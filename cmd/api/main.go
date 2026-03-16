package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kw3a/spotted-server/internal/server"
)

func main() {
	log.Println("Loading .env file")
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading.env file. Running without env variables")
	}
	port := os.Getenv("PORT")
	log.Println("PORT: ", port)
	log.Fatal(server.Run())
}
