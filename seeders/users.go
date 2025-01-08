package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *SeedersConfig) seedUsers() ([]string, error) {
	IDs := []string{
		uuid.New().String(),
	}
	pass, err := bcrypt.GenerateFromPassword([]byte("OhHellYes!"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passHashed := string(pass)
	err = cfg.DB.SeedUser(cfg.Ctx, database.SeedUserParams{
		ID:       IDs[0],
		Name:     "Test User",
		Email:    "myemail@gmail.com",
		Password: passHashed,
		Role:     "dev",
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Users seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/user.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}
