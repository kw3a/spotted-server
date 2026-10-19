package main

import (
	"fmt"
	"regexp"
)

var runIDPattern = regexp.MustCompile(`^[a-f0-9]{16}$`)

func runPrefix(runID string) (string, error) {
	if !runIDPattern.MatchString(runID) {
		return "", fmt.Errorf("run-id debe tener 16 caracteres hexadecimales")
	}
	return "ck_" + runID + "_", nil
}
func runName(runID string) string { return "Capacity run " + runID }
