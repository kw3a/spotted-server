package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

const directoryPath = "test_cases"

func (cfg *SeedersConfig) seedTestCases(problemIDs []string) ([]string, error) {
	problemsQuantity, err := countFiles(directoryPath, "problem*", Directory)
	if err != nil {
		return []string{}, err
	}
	if problemsQuantity != len(problemIDs) {
		return []string{}, errors.New("number of problems and test cases do not match")
	}
	IDs := []string{}
	for i := 1; i <= problemsQuantity; i++ {
		problemPath := fmt.Sprintf("%s/problem%v", directoryPath, i)
		problemIDs, err := cfg.seedTestCasesOnSingleProblem(problemPath, problemIDs[i-1])
		if err != nil {
			return []string{}, err
		}
		IDs = append(IDs, problemIDs...)
	}
	fmt.Println("Test cases seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/test_case.txt"); err != nil {
		return []string{}, err
	}
	return IDs, nil
}

func (cfg *SeedersConfig) seedTestCasesOnSingleProblem(problemPath, problemID string) ([]string, error) {
	testCasesQuantity, err := countFiles(problemPath, "test_case*", Directory)
	if err != nil {
		return []string{}, err
	}
	IDs := []string{}

	for i := 1; i <= testCasesQuantity; i++ {
		ID := uuid.New().String()
		IDs = append(IDs, ID)
		examplePath := fmt.Sprintf("%s/test_case%v", problemPath, i)
		err = cfg.seedTestCase(examplePath, ID, problemID)
		if err != nil {
			return []string{}, err
		}
	}
	return IDs, nil
}

func (cfg *SeedersConfig) seedTestCase(path, ID, problemID string) error {
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
	err = cfg.DB.SeedTestCase(cfg.Ctx, database.SeedTestCaseParams{
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
