package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

const problemsPath = "problems"

func (cfg *SeedersConfig) seedProblems(quizIDs []string) ([]string, error) {
	description1, err := getText(problemsPath + "/description1.seed")
	if err != nil {
		return []string{}, err
	}
	description2, err := getText(problemsPath + "/description2.seed")
	if err != nil {
		return []string{}, err
	}
	problems := []database.SeedProblemParams{
		{
			ID:          uuid.New().String(),
			Description: description1,
			Title:       "Two Sum",
			MemoryLimit: 262144,
			TimeLimit:   1,
			QuizID:      quizIDs[0],
		},
		{
			ID:          uuid.New().String(),
			Description: description2,
			Title:       "Add Two Numbers",
			MemoryLimit: 262144,
			TimeLimit:   1,
			QuizID:      quizIDs[0],
		},
	}
	for _, problem := range problems {
		err = cfg.DB.SeedProblem(cfg.Ctx, problem)
		if err != nil {
			return []string{}, err
		}
	}
	IDs := []string{}
	for _, problem := range problems {
		IDs = append(IDs, problem.ID)
	}
	fmt.Printf("Problems seeded successfully\n")
	if err := SaveIDsToFile(IDs, "to_delete/problem.txt"); err != nil {
		return []string{}, err
	}
	return IDs, nil
}
