package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func seedLanguageProblem(db *sql.DB, languageIDs []int, problemIDs []string) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}
	queries := []string{
		fmt.Sprintf("DELETE FROM language_problem WHERE problem_id = '%s' or problem_id = '%s'", problemIDs[0], problemIDs[1]),
		fmt.Sprintf("INSERT INTO language_problem (id, language_id, problem_id) VALUES ('%s', %d, '%s')", IDs[0], languageIDs[0], problemIDs[0]),
		fmt.Sprintf("INSERT INTO language_problem (id, language_id, problem_id) VALUES ('%s', %d, '%s')", IDs[1], languageIDs[1], problemIDs[0]),
		fmt.Sprintf("INSERT INTO language_problem (id, language_id, problem_id) VALUES ('%s', %d, '%s')", IDs[2], languageIDs[0], problemIDs[1]),
	}

	return IDs, Seed(db, queries, "language_problem")
}
