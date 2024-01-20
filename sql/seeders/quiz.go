package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func seedQuizes(db *sql.DB) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	queries := []string{
		"DELETE FROM quiz WHERE title = 'tech interview 1' or title = 'tech interview 2'",
		fmt.Sprintf("INSERT INTO quiz (id, title, description) VALUES ('%s', 'tech interview 1', 'We need new go developers')", IDs[0]),
		fmt.Sprintf("INSERT INTO quiz (id, title, description) VALUES ('%s', 'tech interview 2', 'We need new java developers')", IDs[1]),
	}

	return IDs, Seed(db, queries, "quizes")
}
