package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

const rootPath = "examples"

func (cfg *SeedersConfig) seedExamples(problemIDs []string) ([]string, error) {
	problemsQuantity, err := countFiles(rootPath, "problem*", Directory)
	if err != nil {
		return []string{}, err
	}
	if problemsQuantity != len(problemIDs) {
		return []string{}, errors.New("number of problems and examples do not match")
	}
	IDs := []string{}
	for i := 1; i <= problemsQuantity; i++ {
		examplesPath := fmt.Sprintf("%s/problem%v", rootPath, i)
		problemIDs, err := cfg.seedExamplesOnSingleProblem(examplesPath, problemIDs[i-1])
		if err != nil {
			return []string{}, err
		}
		IDs = append(IDs, problemIDs...)
	}
	fmt.Println("Examples seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/example.txt"); err != nil {
		return []string{}, err
	}
	return IDs, nil
}

func (cfg *SeedersConfig) seedExamplesOnSingleProblem(problemPath, problemID string) ([]string, error) {
	examplesQuantity, err := countFiles(problemPath, "example*", Directory)
	if err != nil {
		return []string{}, err
	}
	IDs := []string{}

	for i := 1; i <= examplesQuantity; i++ {
		ID := uuid.New().String()
		IDs = append(IDs, ID)
		examplePath := fmt.Sprintf("%s/example%v", problemPath, i)
		err = cfg.seedExample(examplePath, ID, problemID)
		if err != nil {
			return []string{}, err
		}
	}
	return IDs, nil
}

func (cfg *SeedersConfig) seedExample(path, ID, problemID string) error {
	inputPath := fmt.Sprintf("%s/input.seed", path)
	outputPath := fmt.Sprintf("%s/output.seed", path)
	input, err := getText(inputPath)
	if err != nil {
		return err
	}
	output, err := getText(outputPath)
	if err != nil {
		return err
	}
	err = cfg.DB.SeedExample(cfg.Ctx, database.SeedExampleParams{
		ID:        ID,
		Input:     input,
		Output:    output,
		ProblemID: problemID,
	})
	if err != nil {
		return err
	}
	return nil
}
