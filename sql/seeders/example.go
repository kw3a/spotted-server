package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func seedExamples(db *sql.DB, problemIDs []string) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	queries := []string{
		fmt.Sprintf("DELETE FROM example WHERE problem_id = '%s'", problemIDs[0]),
		fmt.Sprintf("INSERT INTO example (id, input, output, problem_id) VALUES ('%s', 'input 1', 'output 1', '%s')", IDs[0], problemIDs[0]),
		fmt.Sprintf("INSERT INTO example (id, input, output, problem_id) VALUES ('%s', 'input 2', 'output 2', '%s')", IDs[1], problemIDs[0]),
	}

	return IDs, Seed(db, queries, "quizes")
}
