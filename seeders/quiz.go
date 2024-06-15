package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedQuizes() ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	for i, id := range IDs {
		err := cfg.DB.SeedQuiz(cfg.Ctx, database.SeedQuizParams{
			ID:          id,
			Title:       fmt.Sprintf("Tech interview %v", i),
			Description: "We are hiring",
			Duration:    120,
		})
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("Quizzes seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/quiz.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}
