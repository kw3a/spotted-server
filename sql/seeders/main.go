package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	source := os.Getenv("DATABASE_URL")
	if source == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := sql.Open("mysql", source)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := seedUsers(db); err != nil {
		log.Fatal(err)
	}
	quizesID, err := seedQuizes(db)
	if err != nil {
		log.Fatal(err)
	}
	languageIDs, err := seedLanguages(db)
	if err != nil {
		log.Fatal(err)
	}
	problemIDs, err := seedProblems(db, quizesID)
	if err != nil {
		log.Fatal(err)
	}
	_, err = seedLanguageProblem(db, languageIDs, problemIDs)
	if err != nil {
		log.Fatal(err)
	}
	_, err = seedExamples(db, problemIDs)
	if err != nil {
		log.Fatal(err)
	}
}

func Seed(db *sql.DB, queries []string, table_name string) error {
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return (err)
		}
	}
	fmt.Printf("Seed %s successfully\n", table_name)
	return nil
}
