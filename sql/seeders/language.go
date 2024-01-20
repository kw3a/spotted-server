package main

import (
	"database/sql"
	"fmt"
)

func seedLanguages(db *sql.DB) ([]int, error) {
	IDs := []int{
		70,
		76,
	}
	queries := []string{
		"DELETE FROM language WHERE id = 70 or id = 76",
		fmt.Sprintf("INSERT INTO language (id, name, version) VALUES (%d, 'Python', 3.9)", IDs[0]),
		fmt.Sprintf("INSERT INTO language (id, name, version) VALUES (%d, 'C++', 14.0)", IDs[1]),
	}

	return IDs, Seed(db, queries, "languages")
}
