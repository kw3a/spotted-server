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
		uuid.New().String(),
		uuid.New().String(),
	}
	pass1, err := bcrypt.GenerateFromPassword([]byte("pass1+x"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passHashed1 := string(pass1)
	err = cfg.DB.SeedUser(cfg.Ctx, database.SeedUserParams{
		ID:       IDs[0],
		Name:     "Braintrust",
		Email:    "contact@braintrust.com",
		Password: passHashed1,
		Role:     "ev",
	})
	if err != nil {
		return nil, err
	}
	pass2, err := bcrypt.GenerateFromPassword([]byte("pass2+x"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passHashed2 := string(pass2)
	err = cfg.DB.SeedUser(cfg.Ctx, database.SeedUserParams{
		ID:       IDs[1],
		Name:     "Launchpad Technologies Inc.",
		Email:    "contact@launch.io",
		Password: passHashed2,
		Role:     "ev",
	})
	if err != nil {
		return nil, err
	}
	pass, err := bcrypt.GenerateFromPassword([]byte("OhHellYes!"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passHashed := string(pass)
	err = cfg.DB.SeedUser(cfg.Ctx, database.SeedUserParams{
		ID:       IDs[2],
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
