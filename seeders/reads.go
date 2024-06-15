package main

import (
	"os"
	"path/filepath"
)

type elementType int

const (
	File elementType = iota
	Directory
	Both
)

func getText(path string) (string, error) {
	path = filepath.Clean(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func countFiles(path, pattern string, elmType elementType) (int, error) {
	count := 0
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		baseName := filepath.Base(filePath)
		matched, err := filepath.Match(pattern, baseName)
		if err != nil {
			return err
		}
		switch elmType {
		case File:
			if matched && !info.IsDir() {
				count++
			}
		case Directory:
			if matched && info.IsDir() {
				count++
			}
		case Both:
			if matched {
				count++
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
