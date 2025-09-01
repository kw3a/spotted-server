package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

func (cfg *SeedersConfig) seedLanguages() ([]int, error) {
	ctx := context.Background()
	languages := []database.SeedLanguageParams{
		{
			ID:          71,
			Name:        "python",
			DisplayName: "Python (3.8.1)",
		},
		{
			ID:          54,
			Name:        "cpp",
			DisplayName: "C++ (GCC 9.2.0)",
		},
		{
			ID:          60,
			Name:        "go",
			DisplayName: "Go (1.13.5)",
		},
		{
			ID:          62,
			Name:        "java",
			DisplayName: "Java (OpenJDK 13.0.1)",
		},
		{
			ID:          63,
			Name:        "javascript",
			DisplayName: "JavaScript (Node.js 12.14.0)",
		},
		{
			ID:          73,
			Name:        "rust",
			DisplayName: "Rust (1.40.0)",
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
		intID, err := strconv.Atoi(id)
		if err != nil {
			return err
		}
		intIDs = append(intIDs, shared.IntToInt32(intID))
	}
	return cfg.DB.DeleteLanguages(ctx, intIDs)
}
