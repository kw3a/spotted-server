package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type judgeAPI struct {
	base    string
	headers http.Header
	client  *http.Client
}

func newJudgeAPI(base, configPath string) (*judgeAPI, error) {
	parsed, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return nil, fmt.Errorf("--judge-url debe ser una URL HTTP(S) de Judge0")
	}
	headers := make(http.Header)
	if configPath != "" {
		config, err := godotenv.Read(configPath)
		if err != nil {
			return nil, err
		}
		for _, auth := range [][3]string{{"AUTHN_HEADER", "AUTHN_TOKEN", "X-Auth-Token"}, {"AUTHZ_HEADER", "AUTHZ_TOKEN", "X-Auth-User"}} {
			if token := config[auth[1]]; token != "" {
				name := config[auth[0]]
				if name == "" {
					name = auth[2]
				}
				headers.Set(name, token)
			}
		}
	}
	return &judgeAPI{parsed.String(), headers, &http.Client{Timeout: 15 * time.Second}}, nil
}

func (api *judgeAPI) request(ctx context.Context, method, path string, body, result any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, api.base+path, reader)
	if err != nil {
		return err
	}
	request.Header = api.headers.Clone()
	request.Header.Set("Content-Type", "application/json")
	response, err := api.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 2000))
		return fmt.Errorf("Judge0 %s %s: HTTP %d: %s", method, strings.Split(path, "?")[0], response.StatusCode, data)
	}
	if result == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(result)
}

type judgeStatus struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}
type judgeResult struct {
	Token         string      `json:"token"`
	Status        judgeStatus `json:"status"`
	Stdout        string      `json:"stdout"`
	Stderr        string      `json:"stderr"`
	CompileOutput string      `json:"compile_output"`
	Message       string      `json:"message"`
	Time          string      `json:"time"`
	Memory        int         `json:"memory"`
}
type judgeCheck struct {
	Algorithm      string `json:"algorithm"`
	LanguageID     string `json:"languageID"`
	Kind           string `json:"kind"`
	SampleKind     string `json:"sampleKind"`
	Case           int    `json:"case"`
	ExpectedStatus int    `json:"expectedStatus"`
	ExpectedOutput string `json:"expectedOutput"`
	Matched        bool   `json:"matched"`
	judgeResult
}
type judgeReport struct {
	Version        string       `json:"judgeVersion"`
	DatasetVersion int          `json:"datasetVersion"`
	URL            string       `json:"judgeURL"`
	VerifiedAt     time.Time    `json:"verifiedAt"`
	CPUTimeLimit   float64      `json:"cpuTimeLimitSeconds"`
	MemoryLimit    int          `json:"memoryLimitKB"`
	Total          int          `json:"total"`
	Matched        int          `json:"matched"`
	Checks         []judgeCheck `json:"checks"`
}

// Las entradas y salidas salen del corpus fijo: no se leen fixtures ni MySQL.
func judgeChecks() []struct {
	check  judgeCheck
	sample Sample
	source string
} {
	var checks []struct {
		check  judgeCheck
		sample Sample
		source string
	}
	for _, languageID := range []string{"54", "62"} {
		for _, algorithm := range algorithms() {
			for _, solution := range algorithm.solutions(languageID) {
				samples := append([]Sample{}, algorithm.Cases...)
				if solution.Kind == "valid" {
					samples = append(samples, algorithm.Examples...)
				}
				for i, sample := range samples {
					status, sampleKind, caseNumber := 3, "secret", i+1
					if i >= len(algorithm.Cases) {
						sampleKind, caseNumber = "example", i-len(algorithm.Cases)+1
					}
					if solution.Kind == "syntax_error" {
						status = 6
					}
					if solution.Kind == "partial" && !algorithm.PartialCaseAccepted[i] {
						status = 4
					}
					check := judgeCheck{Algorithm: algorithm.Key, LanguageID: languageID, Kind: solution.Kind, SampleKind: sampleKind, Case: caseNumber, ExpectedStatus: status, ExpectedOutput: sample.Output}
					checks = append(checks, struct {
						check  judgeCheck
						sample Sample
						source string
					}{check, sample, solution.Source})
				}
			}
		}
	}
	return checks
}

func (api *judgeAPI) waitBatch(ctx context.Context, tokens []string) ([]judgeResult, error) {
	path := "/submissions/batch?base64_encoded=true&tokens=" + strings.Join(tokens, ",") + "&fields=token,status,stdout,stderr,compile_output,message,time,memory"
	for {
		var batch struct {
			Submissions []judgeResult `json:"submissions"`
		}
		if err := api.request(ctx, "GET", path, nil, &batch); err != nil {
			return nil, err
		}
		if len(batch.Submissions) != len(tokens) {
			return nil, fmt.Errorf("Judge0 devolvio un lote incompleto")
		}
		done := true
		for _, result := range batch.Submissions {
			if result.Status.ID <= 2 {
				done = false
			}
		}
		if done {
			for i := range batch.Submissions {
				r := &batch.Submissions[i]
				for _, field := range []*string{&r.Stdout, &r.Stderr, &r.CompileOutput} {
					if *field != "" {
						decoded, err := base64.StdEncoding.DecodeString(*field)
						if err != nil {
							return nil, err
						}
						*field = string(decoded)
					}
				}
			}
			return batch.Submissions, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
}

func judgeResultMatches(check judgeCheck, result judgeResult) bool {
	if result.Status.ID != check.ExpectedStatus {
		return false
	}
	// Accepted debe incluir la salida exacta, no solo una comparacion tolerante.
	if check.ExpectedStatus == 3 {
		return result.Stdout == check.ExpectedOutput
	}
	if check.ExpectedStatus == 4 {
		return result.Stdout != check.ExpectedOutput
	}
	return check.ExpectedStatus == 6 && result.CompileOutput != ""
}

func verifyJudge(ctx context.Context, judgeURL, configPath, reportPath string) error {
	api, err := newJudgeAPI(judgeURL, configPath)
	if err != nil {
		return err
	}
	report := judgeReport{URL: api.base, DatasetVersion: datasetVersion, CPUTimeLimit: 1, MemoryLimit: 262144, VerifiedAt: time.Now().UTC()}
	var about struct {
		Version string `json:"version"`
	}
	if err = api.request(ctx, "GET", "/about", nil, &about); err != nil {
		return err
	}
	report.Version = about.Version
	var available []struct {
		ID int `json:"id"`
	}
	if err = api.request(ctx, "GET", "/languages", nil, &available); err != nil {
		return err
	}
	for _, language := range languages {
		found := false
		for _, item := range available {
			if item.ID == language.ID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Judge0 no tiene habilitado %s (%d)", language.Name, language.ID)
		}
	}
	fmt.Printf("Judge0 %s; CPU=1s, memoria=262144KB; casos directos sin Spotted/MySQL/callbacks\n", report.Version)
	checks := judgeChecks()
	report.Total = len(checks)
	// Lotes pequenos: validar comportamiento, sin convertirlo en una prueba de capacidad.
	for start := 0; start < len(checks); start += 7 {
		end := start + 7
		if end > len(checks) {
			end = len(checks)
		}
		var submissions []map[string]any
		for _, job := range checks[start:end] {
			languageID, _ := strconv.Atoi(job.check.LanguageID)
			encode := func(value string) string { return base64.StdEncoding.EncodeToString([]byte(value)) }
			submissions = append(submissions, map[string]any{"language_id": languageID, "source_code": encode(job.source), "stdin": encode(job.sample.Input), "expected_output": encode(job.sample.Output), "cpu_time_limit": report.CPUTimeLimit, "memory_limit": report.MemoryLimit})
		}
		var created []struct {
			Token string `json:"token"`
		}
		if err = api.request(ctx, "POST", "/submissions/batch?base64_encoded=true", map[string]any{"submissions": submissions}, &created); err != nil {
			break
		}
		if len(created) != end-start {
			err = fmt.Errorf("Judge0 no creo todos los envios")
			break
		}
		tokens := make([]string, len(created))
		for i, item := range created {
			if item.Token == "" {
				err = fmt.Errorf("token de Judge0 ausente")
				break
			}
			tokens[i] = item.Token
		}
		if err != nil {
			break
		}
		var results []judgeResult
		results, err = api.waitBatch(ctx, tokens)
		if err != nil {
			break
		}
		byToken := map[string]judgeResult{}
		for _, result := range results {
			byToken[result.Token] = result
		}
		for i, job := range checks[start:end] {
			check := job.check
			check.judgeResult = byToken[tokens[i]]
			check.Matched = judgeResultMatches(check, check.judgeResult)
			if check.Matched {
				report.Matched++
			} else {
				fmt.Printf("FALLO %s/%s/%s %s %d: esperado %d, obtenido %d (%s)\n", check.Algorithm, check.LanguageID, check.Kind, check.SampleKind, check.Case, check.ExpectedStatus, check.Status.ID, check.Status.Description)
			}
			report.Checks = append(report.Checks, check)
		}
		fmt.Printf("Verificados %d/%d; coinciden %d\n", len(report.Checks), report.Total, report.Matched)
	}
	writeErr := writeJSON(reportPath, report)
	if writeErr != nil {
		return errors.Join(err, writeErr)
	}
	fmt.Printf("Informe: %s; %d/%d resultados esperados\n", reportPath, report.Matched, report.Total)
	if err != nil {
		return err
	}
	if report.Matched != report.Total {
		return fmt.Errorf("Judge0 no produjo todos los veredictos esperados")
	}
	return nil
}
