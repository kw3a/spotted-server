package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func seedProblems(db *sql.DB, quizIDs []string) ([]string, error) {
	content, err := os.ReadFile("problem1.txt")
	if err != nil {
		return []string{}, err
	}
	strContent := escapeQuotes(string(content))
	ID := uuid.New().String()
	ID2 := uuid.New().String()

	const query = `
		INSERT INTO problem (id, description, title, memory_limit, time_limit, quiz_id)
		VALUES 
		('%s', ?, 'Ideal Point', 1000, 1, '%s'),
		('%s', 'dexcrition', 'title 2', 1000, 1, '%s');
	`

	deleteQuery := "DELETE FROM problem WHERE title = 'Ideal Point' or title = 'title 2'"
	_, err = db.Exec(deleteQuery)
	if err != nil {
		return []string{}, err
	}

	insertQuery := fmt.Sprintf(query, ID, quizIDs[0], ID2, quizIDs[0])
	_, err = db.Exec(insertQuery, strContent)
	if err != nil {
		return []string{}, err
	}

	fmt.Printf("Seed problems successfully\n")
	return []string{ID, ID2}, nil
}
