package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOnlyNeedsBaseURLAndNeverUsesDatabase(t *testing.T) {
	for i, database := range []string{"", "DATABASE_URL=invalid-dsn-unused\n"} {
		t.Run([]string{"base-only", "unused-db"}[i], func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "fixtures.json")
			if err := writeJSON(file, manifest()); err != nil {
				t.Fatal(err)
			}
			env := filepath.Join(dir, ".env.b")
			if err := os.WriteFile(env, []byte(database+"BASE_URL=https://B.example.test\n"), 0600); err != nil {
				t.Fatal(err)
			}
			// No ejecuta HTTP real: sustituye solo el proceso k6.
			stub := "#!/bin/sh\nif [ -n \"${DATABASE_URL:-}\" ]; then exit 22; fi\nexit 17\n"
			if err := os.WriteFile(filepath.Join(dir, "k6"), []byte(stub), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("DATABASE_URL", "must-not-reach-k6")
			runs := filepath.Join(dir, "runs")
			err := executeRun([]string{"--env-file", env, "--manifest", file, "--run-dir", runs, "--smoke", "--vus", "1", "--think-scale", "0"})
			var child *exec.ExitError
			if !errors.As(err, &child) || child.ExitCode() != 17 {
				t.Fatalf("runner did not reach k6 / lost child failure: %v", err)
			}
			files, err := filepath.Glob(filepath.Join(runs, "*.json"))
			if err != nil || len(files) != 1 {
				t.Fatalf("missing recovery record: %v %v", files, err)
			}
			data, err := os.ReadFile(files[0])
			if err != nil {
				t.Fatal(err)
			}
			var record RunRecord
			if err = json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			if record.Cleaned || record.FinishedAt.IsZero() || record.Prefix != "ck_"+record.RunID+"_" || strings.Contains(string(data), "DATABASE_URL") || strings.Contains(string(data), "invalid-dsn") {
				t.Fatalf("invalid record / automatic cleanup: %+v", record)
			}
		})
	}
}
