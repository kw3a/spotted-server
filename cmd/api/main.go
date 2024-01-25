package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gitlab.com/kw3a/spotted-server/internal/server"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("port must be set")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Run(portInt))
}
