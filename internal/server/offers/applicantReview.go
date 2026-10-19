package offers

import "github.com/kw3a/spotted-server/internal/server/shared"

// These are presentation models; the storage and HTTP contracts stay unchanged.
type ApplicantReview struct {
	shared.Application
	Problems   []ProblemReview
	Accepted   int
	Total      int
	Solved     int
	Unanswered int
	Percentage int
}

type ProblemReview struct {
	shared.Problem
	Submission shared.Submission
	Cases      []CaseReview
	Accepted   int
	Total      int
}

type CaseReview struct {
	shared.TestCase
	Result shared.TestCaseResult
}

func buildApplicantReviews(applications []shared.Application, problems []shared.Problem) []ApplicantReview {
	reviews := make([]ApplicantReview, 0, len(applications))
	for _, application := range applications {
		review := ApplicantReview{Application: application}
		summaries := make(map[string]shared.Summary, len(application.Summary))
		for _, summary := range application.Summary {
			summaries[summary.Submission.ProblemID] = summary
		}
		for _, problem := range problems {
			summary := summaries[problem.ID]
			item := ProblemReview{Problem: problem, Submission: summary.Submission, Total: len(problem.TestCases)}
			results := make(map[string]shared.TestCaseResult, len(summary.Results))
			for _, result := range summary.Results {
				results[result.TestCaseID] = result
			}
			for _, testCase := range problem.TestCases {
				result := results[testCase.ID]
				item.Cases = append(item.Cases, CaseReview{TestCase: testCase, Result: result})
				if result.Status == "Accepted" && summary.Submission.ID != "" {
					item.Accepted++
				}
			}
			if item.Submission.ID == "" {
				review.Unanswered++
			} else if item.Total > 0 && item.Accepted == item.Total {
				review.Solved++
			}
			review.Accepted += item.Accepted
			review.Total += item.Total
			review.Problems = append(review.Problems, item)
		}
		if review.Total > 0 {
			review.Percentage = review.Accepted * 100 / review.Total
		}
		reviews = append(reviews, review)
	}
	return reviews
}
