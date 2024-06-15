package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DeleteFunction func(context.Context, []string) error

func (cfg *SeedersConfig) CleanDatabase() error {
	type TableData struct {
		FileName string
		Function DeleteFunction
	}
	order := []TableData{
		{
			FileName: "quiz",
			Function: cfg.DB.DeleteQuizes,
		}, {
			FileName: "user",
			Function: cfg.DB.DeleteUsers,
		}, {
			FileName: "language",
			Function: cfg.DeleteLanguagesStr,
		}, {
			FileName: "problem",
			Function: cfg.DB.DeleteProblems,
		}, {
			FileName: "language_problem",
			Function: cfg.DB.DeleteLanguageQuiz,
		}, {
			FileName: "example",
			Function: cfg.DB.DeleteExamples,
		}, {
			FileName: "test_case",
			Function: cfg.DB.DeleteTestCases,
		},
	}
	toDeletePath := "to_delete"
	for _, tableName := range order {
		path := fmt.Sprintf("%s/%s.txt", toDeletePath, tableName.FileName)
		if err := ReadDeleteAndClean(cfg.Ctx, tableName.Function, path); err != nil {
			return err
		}
	}
	return nil
}

func ReadDeleteAndClean(ctx context.Context, fn DeleteFunction, path string) error {
	IDs, err := ReadIDsFromFile(path)
	if err != nil {
		return err
	}
	err = fn(ctx, IDs)
	if err != nil {
		return err
	}
	if err := CleanFile(path); err != nil {
		return err
	}
	return nil
}

func SaveIDsToFile(IDs []string, path string) error {
	path = filepath.Clean(path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, id := range IDs {
		_, err := file.WriteString(id + "\n")
		if err != nil {
			return err
		}
	}

	fmt.Printf("IDs saved to the file '%s'\n", path)
	return nil
}

func ReadIDsFromFile(path string) ([]string, error) {
	path = filepath.Clean(path)
	var IDs []string
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			IDs = append(IDs, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return IDs, nil
}

func CleanFile(path string) error {
	path = filepath.Clean(path)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return err
	}
	fmt.Printf("File content '%s' correctly cleaned.\n", path)
	return nil
}
