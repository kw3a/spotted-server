package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedQuizes(authors []string) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	quizes := []database.SeedQuizParams{
		{
			ID:          IDs[0],
			Duration:    120,
			OfferID:     authors[0],
		},
		{
			ID:          IDs[1],
			Duration:    60,
			OfferID:     authors[1],
		},
	}
	for _, quiz := range quizes {
		err := cfg.DB.SeedQuiz(cfg.Ctx, quiz)
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("Quizes seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/quiz.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}
