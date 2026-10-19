package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Connections struct{ DatabaseURL, BaseURL string }

// Leer con godotenv sin exportar valores ni depender de variables del shell.
func readConnections(path string) (Connections, error) {
	values, err := godotenv.Read(path)
	if err != nil {
		return Connections{}, fmt.Errorf("leer dotenv %s: %w", path, err)
	}
	var extra []string
	for key := range values {
		if key != "DATABASE_URL" && key != "BASE_URL" {
			extra = append(extra, key)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		return Connections{}, fmt.Errorf("dotenv solo admite DATABASE_URL y BASE_URL; mover %s a flags del lanzador", strings.Join(extra, ", "))
	}
	connections := Connections{strings.TrimSpace(values["DATABASE_URL"]), strings.TrimSpace(values["BASE_URL"])}
	return connections, nil
}

type RunSettings struct {
	Mode              string  `json:"TEST_MODE"`
	Smoke             bool    `json:"SMOKE"`
	VUs               int     `json:"MAX_VUS"`
	Ramp              string  `json:"RAMP"`
	Hold              string  `json:"HOLD"`
	RampDown          string  `json:"RAMP_DOWN"`
	Grace             string  `json:"GRACE"`
	ThinkScale        float64 `json:"THINK_SCALE"`
	FilterProbability float64 `json:"P_FILTER"`
	ApplyProbability  float64 `json:"P_APPLY"`
	FinishProbability float64 `json:"P_FINISH"`
	Submissions       int     `json:"SUBMISSIONS_PER_PROBLEM"`
	Results           bool    `json:"ENABLE_RESULTS"`
	Keystrokes        bool    `json:"ENABLE_KEYSTROKES"`
	ResultsTimeout    string  `json:"RESULTS_TIMEOUT"`
	ResultsP95        float64 `json:"RESULTS_P95_MS"`
}

type RuntimeConfig struct {
	RunSettings
	BaseURL      string `json:"BASE_URL"`
	RunID        string `json:"RUN_ID"`
	FixturesFile string `json:"FIXTURES_FILE"`
	EnvFile      string `json:"envFile,omitempty"`
}

type outputs []string

func (o *outputs) String() string { return strings.Join(*o, ",") }
func (o *outputs) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("--out no puede estar vacio")
	}
	*o = append(*o, value)
	return nil
}

type runFlags struct {
	settings       RunSettings
	output         outputs
	summary        string
	quiet, noColor bool
}

func (r *runFlags) register(f *flag.FlagSet) {
	s := &r.settings
	f.StringVar(&s.Mode, "mode", "full", "full o browse")
	f.BoolVar(&s.Smoke, "smoke", false, "un recorrido por VU sin rampas")
	f.IntVar(&s.VUs, "vus", 110, "numero maximo de VUs (1..100000)")
	f.StringVar(&s.Ramp, "ramp", "2m", "duracion de subida")
	f.StringVar(&s.Hold, "hold", "5m", "duracion de carga estable")
	f.StringVar(&s.RampDown, "ramp-down", "30s", "duracion de bajada")
	f.StringVar(&s.Grace, "grace", "", "tiempo para completar recorridos; default 12m en full/30s en browse")
	f.Float64Var(&s.ThinkScale, "think-scale", 1, "multiplicador de pausas (0..100)")
	f.Float64Var(&s.FilterProbability, "p-filter", 0.6, "probabilidad de buscar con filtros (0..1)")
	f.Float64Var(&s.ApplyProbability, "p-apply", 1, "probabilidad de participar (0..1)")
	f.Float64Var(&s.FinishProbability, "p-finish", 1, "probabilidad de finalizar (0..1)")
	f.IntVar(&s.Submissions, "submissions-per-problem", 3, "intentos por problema (3..30)")
	f.BoolVar(&s.Results, "enable-results", true, "verificar veredictos SSE; obligatorio en full")
	f.BoolVar(&s.Keystrokes, "enable-keystrokes", true, "enviar ventanas de pulsaciones")
	f.StringVar(&s.ResultsTimeout, "results-timeout", "65s", "timeout de la conexion SSE")
	f.Float64Var(&s.ResultsP95, "results-p95-ms", 60000, "limite de p95 del juez en milisegundos")
	f.Var(&r.output, "out", "salida k6, por ejemplo json=b.results.json; repetible")
	f.StringVar(&r.summary, "summary-export", "", "archivo de resumen JSON")
	f.BoolVar(&r.quiet, "quiet", false, "ocultar progreso k6")
	f.BoolVar(&r.noColor, "no-color", false, "desactivar colores k6")
}

func (r *runFlags) resolve(f *flag.FlagSet) {
	explicit := map[string]bool{}
	f.Visit(func(option *flag.Flag) { explicit[option.Name] = true })
	if r.settings.Mode == "browse" {
		if !explicit["enable-results"] {
			r.settings.Results = false
		}
		if !explicit["enable-keystrokes"] {
			r.settings.Keystrokes = false
		}
	}
	if r.settings.Grace == "" && !explicit["grace"] {
		r.settings.Grace = "12m"
		if r.settings.Mode == "browse" {
			r.settings.Grace = "30s"
		}
	}
}

func (s RunSettings) validate(baseURL string) error {
	if err := validateBaseURL(baseURL); err != nil {
		return err
	}
	if s.Mode != "full" && s.Mode != "browse" {
		return fmt.Errorf("--mode debe ser full o browse")
	}
	if s.VUs < 1 || s.VUs > 100000 {
		return fmt.Errorf("--vus debe estar entre 1 y 100000")
	}
	if s.Submissions < 3 || s.Submissions > 30 {
		return fmt.Errorf("--submissions-per-problem debe estar entre 3 y 30")
	}
	for _, item := range []struct{ name, value string }{{"ramp", s.Ramp}, {"hold", s.Hold}, {"ramp-down", s.RampDown}, {"grace", s.Grace}, {"results-timeout", s.ResultsTimeout}} {
		value, err := time.ParseDuration(item.value)
		if err != nil || value <= 0 {
			return fmt.Errorf("--%s debe ser una duracion positiva", item.name)
		}
	}
	for _, item := range []struct {
		name            string
		value, min, max float64
	}{{"think-scale", s.ThinkScale, 0, 100}, {"p-filter", s.FilterProbability, 0, 1}, {"p-apply", s.ApplyProbability, 0, 1}, {"p-finish", s.FinishProbability, 0, 1}, {"results-p95-ms", s.ResultsP95, 1, 3600000}} {
		if math.IsNaN(item.value) || math.IsInf(item.value, 0) || item.value < item.min || item.value > item.max {
			return fmt.Errorf("--%s fuera de rango", item.name)
		}
	}
	if s.Mode == "full" && (!s.Results || s.ApplyProbability == 0 || s.FinishProbability == 0) {
		return fmt.Errorf("full requiere resultados, participacion y finalizacion")
	}
	if s.Mode == "browse" && (s.Results || s.Keystrokes) {
		return fmt.Errorf("browse no admite resultados ni pulsaciones")
	}
	return nil
}

func validateBaseURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("BASE_URL debe apuntar por HTTPS al servidor externo, sin rutas ni credenciales")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "127" || strings.HasPrefix(host, "127.") || host == "0.0.0.0" || host == "::" || host == "::1" {
		return fmt.Errorf("BASE_URL requiere servidor externo, no loopback")
	}
	return nil
}

func (r runFlags) k6Args() []string {
	var args []string
	if r.summary != "" {
		args = append(args, "--summary-export", r.summary)
	}
	for _, output := range r.output {
		args = append(args, "--out", output)
	}
	if r.quiet {
		args = append(args, "--quiet")
	}
	if r.noColor {
		args = append(args, "--no-color")
	}
	return args
}

// La configuracion es un modulo privado por ejecucion; no usa -e ni __ENV.
func createLaunch(config RuntimeConfig, directory, mainFile string) (path string, release func(), err error) {
	source, err := json.Marshal(config)
	if err != nil {
		return "", nil, err
	}
	if err = os.MkdirAll(directory, 0700); err != nil {
		return "", nil, err
	}
	temporary, err := os.MkdirTemp(directory, "run-")
	if err != nil {
		return "", nil, err
	}
	release = func() { _ = os.RemoveAll(temporary) }
	mainPath, err := filepath.Abs(mainFile)
	if err != nil {
		release()
		return "", nil, err
	}
	quotedPath, _ := json.Marshal(filepath.ToSlash(mainPath))
	parameters := append([]byte("globalThis.__CAPACITY_CONFIG = "), source...)
	parameters = append(parameters, []byte(";\n")...)
	if err = os.WriteFile(filepath.Join(temporary, "parameters.js"), parameters, 0600); err != nil {
		release()
		return "", nil, err
	}
	path = filepath.Join(temporary, "main.js")
	entry := "import './parameters.js';\nexport { options, setup, applicant } from " + string(quotedPath) + ";\n"
	if err = os.WriteFile(path, []byte(entry), 0600); err != nil {
		release()
		return "", nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		release()
		return "", nil, err
	}
	return absolute, release, nil
}
