package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

type SeedersConfig struct {
	DB  *database.Queries
	Ctx context.Context
}

func main() {
	seedCfg := SeedersConfig{}
	err := godotenv.Load("../.env")
	source := os.Getenv("DATABASE_URL")
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	if source == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := sql.Open("mysql", source)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	seedCfg = SeedersConfig{
		DB:  dbQueries,
		Ctx: context.Background(),
	}
	log.Println("Connected to database!")
	if err := seedCfg.CleanDatabase(); err != nil {
		log.Fatal(err)
	}
	_, err = seedCfg.seedUsers()
	if err != nil {
		log.Fatal(err)
	}
	_, err = seedCfg.seedLanguages()
	if err != nil {
		log.Fatal(err)
	}
	/*
		companies, err := seedCfg.seedCompanies(users)
		if err != nil {
			log.Fatal(err)
		}
		offers, err := seedCfg.seedOffers(companies)
		if err != nil {
			log.Fatal(err)
		}
		quizesID, err := seedCfg.seedQuizes(offers)
		if err != nil {
			log.Fatal(err)
		}
		problemIDs, err := seedCfg.seedProblems(quizesID)
		if err != nil {
			log.Fatal(err)
		}
		languageProblems := []database.SeedLanguageQuizParams{}
		for _, langID := range languageIDs {
			languageProblems = append(languageProblems, database.SeedLanguageQuizParams{
				ID:         uuid.New().String(),
				LanguageID: shared.IntToInt32(langID),
				QuizID:     quizesID[0],
			})
		}
		languageProblems = append(languageProblems, database.SeedLanguageQuizParams{
			ID:         uuid.New().String(),
			LanguageID: shared.IntToInt32(languageIDs[0]),
			QuizID:     quizesID[1],
		})

		_, err = seedCfg.seedLanguageProblem(languageProblems)
		if err != nil {
			log.Fatal(err)
		}
		_, err = seedCfg.seedExamples(problemIDs)
		if err != nil {
			log.Fatal(err)
		}
		_, err = seedCfg.seedTestCases(problemIDs)
		if err != nil {
			log.Fatal(err)
		}
	*/
}
