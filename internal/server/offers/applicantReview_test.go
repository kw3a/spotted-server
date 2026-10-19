package offers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

func TestApplicantReviewIncludesUnansweredProblems(t *testing.T) {
	problems := []shared.Problem{
		{ID: "p1", TestCases: []shared.TestCase{{ID: "c1"}, {ID: "c2"}}},
		{ID: "p2", TestCases: []shared.TestCase{{ID: "c3"}}},
	}
	applications := []shared.Application{{Summary: []shared.Summary{{
		Submission: shared.Submission{ID: "submission", ProblemID: "p1"},
		Results: []shared.TestCaseResult{
			{TestCaseID: "c2", Status: "Wrong Answer", Output: "obtained"},
			{TestCaseID: "c1", Status: "Accepted"},
		},
	}}}}
	review := buildApplicantReviews(applications, problems)[0]
	if review.Accepted != 1 || review.Total != 3 || review.Percentage != 33 || review.Solved != 0 || review.Unanswered != 1 {
		t.Fatalf("incorrect overall result: %+v", review)
	}
	if len(review.Problems) != 2 || review.Problems[1].Submission.ID != "" {
		t.Fatal("unanswered problems must remain visible")
	}
	if review.Problems[0].Cases[1].Result.Output != "obtained" {
		t.Fatal("results must be matched by case ID, not their query order")
	}
}

func TestApplicantReviewZeroCasesIsNotSolved(t *testing.T) {
	review := buildApplicantReviews([]shared.Application{{Summary: []shared.Summary{{Submission: shared.Submission{ID: "s", ProblemID: "p"}}}}}, []shared.Problem{{ID: "p"}})[0]
	if review.Solved != 0 || review.Percentage != 0 || review.Unanswered != 0 {
		t.Fatalf("zero cases must not imply a perfect result: %+v", review)
	}
}

func TestApplicantReviewTemplate(t *testing.T) {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"inc": func(a int) int { return a + 1 },
		"json": func(value interface{}) template.JS {
			encoded, _ := json.Marshal(value)
			return template.JS(encoded)
		},
	}
	templates, err := template.New("").Funcs(funcs).ParseGlob("../../../views/*.html")
	if err != nil {
		t.Fatal(err)
	}
	problems := []shared.Problem{{ID: "p", Title: "Problem", TestCases: []shared.TestCase{{ID: "c", Input: "1", Output: "2"}}}}
	applications := []shared.Application{{Applicant: shared.User{Name: "Test Applicant"}, Participation: shared.Participation{ID: "participation"}, Summary: []shared.Summary{{Submission: shared.Submission{ID: "s", ProblemID: "p", Src: "</code><script>alert(1)</script>"}}}}}
	for _, reviews := range [][]ApplicantReview{buildApplicantReviews(applications, problems), nil} {
		var output bytes.Buffer
		if err := templates.ExecuteTemplate(&output, "offerAdmin", ApplicantsData{Applicants: reviews, Problems: problems}); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(output.String(), "<script>alert(1)</script>") {
			t.Fatal("submitted code must be escaped")
		}
		if !strings.Contains(output.String(), "candidate-list") {
			t.Fatal("missing review layout")
		}
	}
}
