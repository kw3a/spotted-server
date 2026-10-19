package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// k6 inspect evalua solo la inicializacion: no ejecuta setup ni VUs ni HTTP.
func TestK6InspectionWithoutServices(t *testing.T) {
	if _, err := exec.LookPath("k6"); err != nil {
		t.Skip("k6 required")
	}
	file := filepath.Join(t.TempDir(), "fixtures.json")
	if err := writeJSON(file, manifest()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	settings := testRunSettings(t)
	settings.Smoke, settings.VUs = true, 1
	entry, release, err := createLaunch(RuntimeConfig{RunSettings: settings, BaseURL: "https://capacity.example.test", RunID: "0123456789abcdef", FixturesFile: file}, t.TempDir(), "../main.js")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	command := exec.CommandContext(ctx, "k6", "inspect", "--execution-requirements", entry)
	command.Dir = ".."
	// Valores antiguos no deben sustituir los parametros del modulo generado.
	command.Env = append(os.Environ(), "K6_NO_USAGE_REPORT=true", "SMOKE=0", "MAX_VUS=999", "BASE_URL=https://localhost")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("k6 initialization failed: %s (%v)", output, err)
	}
	var options struct {
		Scenarios map[string]struct {
			Executor string `json:"executor"`
			VUs      int    `json:"vus"`
		} `json:"scenarios"`
	}
	if err = json.Unmarshal(output, &options); err != nil {
		t.Fatal(err)
	}
	if options.Scenarios["applicant"].Executor != "per-vu-iterations" || options.Scenarios["applicant"].VUs != 1 {
		t.Fatalf("launcher flags were overridden: %s", output)
	}
	for _, language := range []string{"cpp", "java"} {
		if !strings.Contains(string(output), "submissions_verified{language:"+language+"}") {
			t.Fatalf("missing coverage for %s: %s", language, output)
		}
	}
}
