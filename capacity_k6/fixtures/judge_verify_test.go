package main

import "testing"

func TestJudgeCheckPlan(t *testing.T) {
	checks := judgeChecks()
	if len(checks) != 102 {
		t.Fatalf("got %d checks, want 102", len(checks))
	}
	statuses, languages, categories := map[int]int{}, map[string]int{}, map[string]int{}
	for _, job := range checks {
		statuses[job.check.ExpectedStatus]++
		languages[job.check.LanguageID]++
		categories[job.check.SampleKind]++
		if job.source == "" || job.sample.Input == "" || job.sample.Output != job.check.ExpectedOutput {
			t.Fatal("incomplete direct payload")
		}
	}
	if statuses[3] != 56 || statuses[4] != 16 || statuses[6] != 30 {
		t.Fatalf("unexpected verdict distribution: %v", statuses)
	}
	if languages["54"] != 51 || languages["62"] != 51 || categories["secret"] != 90 || categories["example"] != 12 {
		t.Fatal("incorrect coverage")
	}
}

func TestJudgeResultValidation(t *testing.T) {
	valid := judgeCheck{ExpectedStatus: 3, ExpectedOutput: "4\n"}
	if !judgeResultMatches(valid, judgeResult{Status: judgeStatus{ID: 3}, Stdout: "4\n"}) {
		t.Fatal("rejected valid result")
	}
	if judgeResultMatches(valid, judgeResult{Status: judgeStatus{ID: 3}, Stdout: "4"}) {
		t.Fatal("accepted non-exact output")
	}
	invalid := judgeCheck{ExpectedStatus: 6}
	if !judgeResultMatches(invalid, judgeResult{Status: judgeStatus{ID: 6}, CompileOutput: "syntax error"}) {
		t.Fatal("rejected compilation error")
	}
	if judgeResultMatches(invalid, judgeResult{Status: judgeStatus{ID: 11}, CompileOutput: "sandbox failure"}) {
		t.Fatal("mistook runtime error for compilation error")
	}
	if judgeResultMatches(invalid, judgeResult{Status: judgeStatus{ID: 6}}) {
		t.Fatal("accepted error without compiler diagnostic")
	}
}
