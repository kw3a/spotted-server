package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/kw3a/spotted-server/internal/auth"
)

func seedUsers(db *sql.DB) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
	}
	passHashed1, err := auth.HashPassword("OhHellYes!")
	if err != nil {
		return nil, err
	}
	queries := []string{
		"DELETE FROM user WHERE name = 'user 1'",
		fmt.Sprintf("INSERT INTO user (id, name, email, password) VALUES ('%s', 'user 1', 'myemail@gmail.com', '%s')", IDs[0], passHashed1),
	}
	return IDs, Seed(db, queries, "users")
}
