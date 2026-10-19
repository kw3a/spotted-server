package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestManifestShapeAndSecretSeparation(t *testing.T) {
	m := manifest()
	if len(m.Offers) != 10 {
		t.Fatal("expected 10 offers")
	}
	ids := map[string]bool{}
	languageCounts := map[string]int{}
	for i, o := range m.Offers {
		candidates := []string{o.OfferID, o.CompanyID, o.QuizID, fixtureID(fmt.Sprintf("user/%02d", i+1))}
		if len(o.Problems) != 3 {
			t.Fatal("expected three problems")
		}
		quizLanguages := map[string]bool{}
		for _, p := range o.Problems {
			candidates = append(candidates, p.ProblemID)
			languageCounts[p.LanguageID]++
			quizLanguages[p.LanguageID] = true
			if len(p.Solutions) != 3 {
				t.Fatal("expected valid, partial and syntax-error solutions")
			}
			if p.TestCaseCount != 5 || (p.LanguageID != "54" && p.LanguageID != "62") || p.Source == "" {
				t.Fatal("missing solution or tests")
			}
		}
		if len(quizLanguages) != 2 {
			t.Fatal("each quiz must include C++ and Java")
		}
		for _, id := range candidates {
			if uuid.Validate(id) != nil || ids[id] {
				t.Fatal("invalid or reused ID")
			}
			ids[id] = true
		}
	}
	if languageCounts["54"] != 15 || languageCounts["62"] != 15 {
		t.Fatal("expected 15 problems per language")
	}
	data, _ := json.Marshal(m)
	for _, field := range []string{`"input"`, `"output"`, `"password"`, `"cases"`, `"examples"`} {
		if strings.Contains(string(data), field) {
			t.Fatalf("manifest exposes %s", field)
		}
	}
	again, _ := json.Marshal(manifest())
	if string(data) != string(again) {
		t.Fatal("manifest is not deterministic")
	}
}

type compiledProgram struct {
	executable string
	args       []string
}

func compileSolution(t *testing.T, languageID, source string, syntaxError bool) compiledProgram {
	t.Helper()
	compiler := "g++"
	if languageID == "62" {
		compiler = "javac"
	}
	if _, err := exec.LookPath(compiler); err != nil {
		t.Skip(compiler + " required for offline solution verification")
	}
	if languageID == "62" {
		if _, err := exec.LookPath("java"); err != nil {
			t.Skip("java required")
		}
	}
	directory := t.TempDir()
	filename := filepath.Join(directory, "main.cpp")
	executable := filepath.Join(directory, "solution")
	args := []string{"-std=c++17", "-O2", filename, "-o", executable}
	program := compiledProgram{executable, nil}
	if languageID == "62" {
		filename = filepath.Join(directory, "Main.java")
		args = []string{"--release", "13", "-d", directory, filename}
		program = compiledProgram{"java", []string{"-Xmx128m", "-cp", directory, "Main"}}
	}
	if err := os.WriteFile(filename, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, compiler, args...).CombinedOutput()
	if syntaxError {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || ctx.Err() != nil {
			t.Fatalf("expected compilation error, got %s (%v)", output, err)
		}
		return compiledProgram{}
	}
	if err != nil {
		t.Fatalf("compilation failed: %s (%v)", output, err)
	}
	return program
}

func (p compiledProgram) run(t *testing.T, input string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, p.executable, p.args...)
	command.Stdin = strings.NewReader(input)
	output, err := command.Output()
	return string(output), err
}

func TestSolutionsAgainstExamplesAndSecretCases(t *testing.T) {
	for _, languageID := range []string{"54", "62"} {
		t.Run(languageID, func(t *testing.T) {
			for _, a := range algorithms() {
				if len(a.Examples) != 2 || len(a.Cases) != 5 {
					t.Fatal("incorrect sample counts")
				}
				examples := map[string]bool{}
				for _, ex := range a.Examples {
					examples[ex.Input] = true
				}
				for _, sample := range append(append([]Sample{}, a.Examples...), a.Cases...) {
					if len([]byte(sample.Input)) > 65535 || len([]byte(sample.Output)) > 65535 {
						t.Fatal("sample exceeds MySQL TEXT size")
					}
				}
				for _, tc := range a.Cases {
					if examples[tc.Input] {
						t.Fatal("secret case duplicates example")
					}
				}
				for _, solution := range a.solutions(languageID) {
					t.Run(a.Key+"/"+solution.Kind, func(t *testing.T) {
						program := compileSolution(t, languageID, solution.Source, solution.Kind == "syntax_error")
						if solution.Kind == "syntax_error" {
							if solution.Expected != (Verdict{Errors: 5}) {
								t.Fatal("incorrect compilation-error verdict")
							}
							return
						}
						if solution.Kind == "valid" {
							for _, ex := range a.Examples {
								output, err := program.run(t, ex.Input)
								if err != nil || output != ex.Output {
									t.Fatalf("example: got %q (%v), want %q", output, err, ex.Output)
								}
							}
						}
						actual := Verdict{}
						for index, tc := range a.Cases {
							output, err := program.run(t, tc.Input)
							if solution.Kind == "partial" && err == nil && (output == tc.Output) != a.PartialCaseAccepted[index] {
								t.Fatalf("partial verdict differs from direct-Judge0 oracle for case %d", index+1)
							}
							if err != nil {
								actual.Errors++
							} else if output == tc.Output {
								actual.Accepted++
							} else {
								actual.WrongAnswer++
							}
						}
						if actual != solution.Expected {
							t.Fatalf("got %+v want %+v", actual, solution.Expected)
						}
					})
				}
			}
		})
	}
}

func TestCleanupScope(t *testing.T) {
	for _, id := range []string{"", "%", "0123456789abcde_", "0123456789abcdef_", "ABCDEF0123456789"} {
		if _, err := runPrefix(id); err == nil {
			t.Fatalf("accepted unsafe run ID %q", id)
		}
	}
	id := fixtureID("runtime-user")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: id})
	raw, err := token.SignedString([]byte("test-only"))
	if err != nil {
		t.Fatal(err)
	}
	if !tokenBelongsTo(raw, map[string]bool{id: true}) {
		t.Fatal("failed to select own JWT")
	}
	if tokenBelongsTo(raw, map[string]bool{"other": true}) || tokenBelongsTo("invalid", map[string]bool{id: true}) {
		t.Fatal("selected unrelated JWT")
	}
}

// Oraculos independientes para casos pequenos: DP cuadratica, enumeracion
// exhaustiva de subconjuntos y Bellman-Ford, en vez de los algoritmos enviados.
func TestValidSolutionsAgainstIndependentOracles(t *testing.T) {
	random := rand.New(rand.NewSource(42))
	all := algorithms()
	programs := map[string]compiledProgram{}
	for _, languageID := range []string{"54", "62"} {
		for _, a := range all {
			programs[a.Key+"/"+languageID] = compileSolution(t, languageID, programsFor(a.Key, languageID).Source, false)
		}
	}
	for trial := 0; trial < 20; trial++ {
		n := 1 + random.Intn(12)
		values, longest := make([]int, n), make([]int, n)
		answer := 0
		for i := range values {
			values[i] = random.Intn(11) - 5
			longest[i] = 1
			for j := 0; j < i; j++ {
				if values[j] < values[i] && longest[j]+1 > longest[i] {
					longest[i] = longest[j] + 1
				}
			}
			if longest[i] > answer {
				answer = longest[i]
			}
		}
		cases := []Sample{sequenceSample(values, answer)}
		capacity := 1 + random.Intn(20)
		weights, profits := make([]int, n), make([]int, n)
		var bag strings.Builder
		fmt.Fprintf(&bag, "%d %d\n", n, capacity)
		for i := 0; i < n; i++ {
			weights[i], profits[i] = 1+random.Intn(12), 1+random.Intn(30)
			fmt.Fprintf(&bag, "%d %d\n", weights[i], profits[i])
		}
		answer = 0
		for mask := 0; mask < (1 << n); mask++ {
			weight, profit := 0, 0
			for i := 0; i < n; i++ {
				if mask&(1<<i) != 0 {
					weight += weights[i]
					profit += profits[i]
				}
			}
			if weight <= capacity && profit > answer {
				answer = profit
			}
		}
		cases = append(cases, Sample{bag.String(), fmt.Sprintf("%d\n", answer)})
		vertices, edgeCount := 1+random.Intn(8), random.Intn(30)
		type edge struct{ u, v, weight int }
		edges := make([]edge, edgeCount)
		distance := make([]int, vertices)
		const infinity = 1 << 30
		for i := 1; i < vertices; i++ {
			distance[i] = infinity
		}
		var graph strings.Builder
		fmt.Fprintf(&graph, "%d %d\n", vertices, edgeCount)
		for i := range edges {
			edges[i] = edge{random.Intn(vertices), random.Intn(vertices), random.Intn(20)}
			e := edges[i]
			fmt.Fprintf(&graph, "%d %d %d\n", e.u+1, e.v+1, e.weight)
		}
		for i := 1; i < vertices; i++ {
			for _, e := range edges {
				if distance[e.u] != infinity && distance[e.u]+e.weight < distance[e.v] {
					distance[e.v] = distance[e.u] + e.weight
				}
			}
		}
		answer = distance[vertices-1]
		if answer == infinity {
			answer = -1
		}
		cases = append(cases, Sample{graph.String(), fmt.Sprintf("%d\n", answer)})
		for i, sample := range cases {
			for _, languageID := range []string{"54", "62"} {
				t.Run(fmt.Sprintf("%s/%s/random-%d", all[i].Key, languageID, trial), func(t *testing.T) {
					output, err := programs[all[i].Key+"/"+languageID].run(t, sample.Input)
					if err != nil || output != sample.Output {
						t.Fatalf("got %q (%v), want %q for input %q", output, err, sample.Output, sample.Input)
					}
				})
			}
		}
	}
}
