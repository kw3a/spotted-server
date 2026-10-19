package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseRunFlags(t *testing.T, args ...string) runFlags {
	t.Helper()
	f := flag.NewFlagSet("run", flag.ContinueOnError)
	var run runFlags
	run.register(f)
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	run.resolve(f)
	return run
}
func testRunSettings(t *testing.T) RunSettings { return parseRunFlags(t).settings }

func TestDotenvOnlyConnectionsAndFileAuthority(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.b")
	t.Setenv("DATABASE_URL", "must-not-be-used")
	t.Setenv("BASE_URL", "https://localhost")
	write := func(text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("DATABASE_URL='user:secret@tcp(B:3306)/db?parseTime=true'\nBASE_URL=https://B.example.test\n")
	config, err := readConnections(path)
	if err != nil || config.BaseURL != "https://B.example.test" || !strings.Contains(config.DatabaseURL, "tcp(B:3306)") {
		t.Fatalf("dotenv was not authoritative: %+v %v", config, err)
	}
	if os.Getenv("DATABASE_URL") != "must-not-be-used" {
		t.Fatal("dotenv changed the process environment")
	}
	write("DATABASE_URL=test\nBASE_URL=https://B.example.test\nMAX_VUS=110\n")
	if _, err = readConnections(path); err == nil || !strings.Contains(err.Error(), "MAX_VUS") {
		t.Fatal("accepted load settings in dotenv")
	}
	if _, err = readConnections(path + ".missing"); err == nil {
		t.Fatal("missing dotenv fell back to shell variables")
	}
}

func TestRunFlagsDefaultsAndModes(t *testing.T) {
	full := parseRunFlags(t)
	if err := full.settings.validate("https://B.example.test"); err != nil {
		t.Fatal(err)
	}
	if full.settings.VUs != 110 || !full.settings.Results || !full.settings.Keystrokes || full.settings.Grace != "12m" {
		t.Fatal("incorrect full defaults")
	}
	browse := parseRunFlags(t, "--mode", "browse")
	if err := browse.settings.validate("https://B.example.test"); err != nil {
		t.Fatal(err)
	}
	if browse.settings.Results || browse.settings.Keystrokes || browse.settings.Grace != "30s" {
		t.Fatal("incorrect browse defaults")
	}
	explicit := parseRunFlags(t, "--mode", "browse", "--enable-results=true")
	if err := explicit.settings.validate("https://B.example.test"); err == nil {
		t.Fatal("accepted incompatible explicit option")
	}
	smoke := parseRunFlags(t, "--smoke", "--vus", "1", "--think-scale", "0", "--out", "json=b.results.json", "--summary-export", "b.summary.json")
	if !smoke.settings.Smoke || smoke.settings.VUs != 1 || smoke.settings.ThinkScale != 0 {
		t.Fatal("CLI flags ignored")
	}
	for _, arg := range smoke.k6Args() {
		if arg == "-e" || arg == "--env" {
			t.Fatal("transported settings as environment")
		}
	}
}

func TestInvalidFlagsStopBeforeDatabase(t *testing.T) {
	for _, args := range [][]string{{"--vus", "0"}, {"--think-scale", "NaN"}, {"--submissions-per-problem", "2"}, {"--grace", ""}, {"--ramp", "0s"}, {"--enable-results=false"}} {
		if err := parseRunFlags(t, args...).settings.validate("https://B.example.test"); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if err := testRunSettings(t).validate("https://127.0.0.1"); err == nil {
		t.Fatal("accepted loopback")
	}
}

func TestLaunchModuleNoSecretsAndIsolatedLifecycle(t *testing.T) {
	settings := testRunSettings(t)
	config := RuntimeConfig{RunSettings: settings, BaseURL: "https://B.example.test", RunID: "0123456789abcdef", FixturesFile: "/tmp/fixtures.json"}
	directory := t.TempDir()
	first, release, err := createLaunch(config, directory, "../main.js")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	second, releaseOther, err := createLaunch(config, directory, "../main.js")
	if err != nil {
		t.Fatal(err)
	}
	defer releaseOther()
	if first == second {
		t.Fatal("concurrent runs share configuration")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(first), "parameters.js"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "DATABASE_URL") || strings.Contains(string(data), "__ENV") {
		t.Fatal("runtime module exposes SQL/environment settings")
	}
	release()
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatal("temporary entry was not removed")
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatal("one run removed another run's module")
	}
}

func TestCLIValidatesBeforeConnecting(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.b")
	if err := os.WriteFile(path, []byte("DATABASE_URL=not-a-dsn\nBASE_URL=https://B.example.test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	err := execute([]string{"run", "--env-file", path, "--vus", "0"})
	if err == nil || !strings.Contains(err.Error(), "--vus") {
		t.Fatalf("did not validate CLI before SQL: %v", err)
	}
}
