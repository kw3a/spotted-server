package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedLanguages() ([]int, error) {
	ctx := context.Background()
	languages := []database.SeedLanguageParams{
		{
			ID:      71,
			Name:    "python",
			Version: 3,
		},
		{
			ID:      54,
			Name:    "cpp",
			Version: 9,
		},
	}

	for _, lang := range languages {
		err := cfg.DB.SeedLanguage(ctx, lang)
		if err != nil {
			return nil, err
		}
	}
	IDs := []int{}
	for _, lang := range languages {
		IDs = append(IDs, int(lang.ID))
	}
	stringIDs := []string{}
	for _, id := range IDs {
		stringIDs = append(stringIDs, strconv.Itoa(id))
	}
	fmt.Println("Languages seeded successfully")
	if err := SaveIDsToFile(stringIDs, "to_delete/language.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}

func (cfg *SeedersConfig) DeleteLanguagesStr(ctx context.Context, IDs []string) error {
	intIDs := []int32{}
	for _, id := range IDs {
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		intIDs = append(intIDs, int32(intID))
	}
	return cfg.DB.DeleteLanguages(ctx, intIDs)
}
