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
		ID:          IDs[0],
		Nick:        "genericBoss",
		Name:        "Beñat Beniju",
		Email:       "myemail@gmail.com",
		Number:      "+591 69546920",
		Password:    passHashed,
		Description: "Usuario de prueba con 2 publicaciones",
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
