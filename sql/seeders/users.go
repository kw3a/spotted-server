package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func seedUsers(db *sql.DB) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	queries := []string{
		"DELETE FROM user WHERE name = 'user 1' or name = 'user 2'",
		fmt.Sprintf("INSERT INTO user (id, name) VALUES ('%s', 'user 1')", IDs[0]),
		fmt.Sprintf("INSERT INTO user (id, name) VALUES ('%s', 'user 2')", IDs[1]),
	}
	return IDs, Seed(db, queries, "users")
}
