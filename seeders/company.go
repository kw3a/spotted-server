package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedCompanies(users []string) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	err := cfg.DB.SeedCompany(cfg.Ctx, database.SeedCompanyParams{
		ID:          IDs[0],
		UserID:      users[0],
		Name:        "Braintrust",
		Description: "We are a team of developers",
		Website:     "https://braintrust.com",
		ImageUrl:    "",
	})
	if err != nil {
		return nil, err
	}
	err = cfg.DB.SeedCompany(cfg.Ctx, database.SeedCompanyParams{
		ID:          IDs[1],
		UserID:      users[0],
		Name:        "Launchpad Technologies Inc.",
		Description: "We are a team of developers",
		Website:     "https://launch.io",
		ImageUrl:    "",
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Companies seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/company.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}
