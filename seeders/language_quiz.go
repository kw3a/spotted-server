package main

import (
	"fmt"

	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedLanguageProblem(quizLanguages []database.SeedLanguageQuizParams) ([]string, error) {
	for _, problem := range quizLanguages {
		err := cfg.DB.SeedLanguageQuiz(cfg.Ctx, problem)
		if err != nil {
			return []string{}, err
		}
	}
	IDs := []string{}
	for _, problem := range quizLanguages {
		IDs = append(IDs, problem.ID)
	}
	fmt.Println("language_quiz seeded succesfully")
	if err := SaveIDsToFile(IDs, "to_delete/language_problem.txt"); err != nil {
		return []string{}, err
	}
	return IDs, nil
}
