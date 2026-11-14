package inc2test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/db_integration_test/testutil"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/require"
)

func TestRegisterOffer(t *testing.T) {
	for _, name := range []string{"success", "invalidLanguage", "duplicateLanguage", "invalidProblem", "invalidTestCase", "invalidExample"} {
		t.Run(name, func(t *testing.T) {
			f := testutil.New(t)
			offerID, quizID := uuid.NewString(), uuid.NewString()
			offer := shared.Offer{Title: "Developer", About: "About", Requirements: "Requirements", Benefits: "Benefits", MinWage: 100, MaxWage: 200, CompanyID: f.CompanyID}
			quiz := shared.Quiz{Duration: 45, Languages: []int32{54, 62}}
			problems := []shared.Problem{
				{Title: "First", Description: "First description", TimeLimit: 1500, TestCases: []shared.TestCase{{Input: "1", Output: "2"}, {Input: "3", Output: "4"}}, Examples: []shared.Example{{Input: "5", Output: "6"}}},
				{Title: "Second", Description: "Second description", TimeLimit: 2500, TestCases: []shared.TestCase{{Input: "7", Output: "8"}}, Examples: []shared.Example{{Input: "9", Output: "10"}}},
			}
			switch name {
			case "invalidLanguage":
				quiz.Languages = []int32{54, 999}
			case "duplicateLanguage":
				quiz.Languages = []int32{54, 54}
			case "invalidProblem":
				problems[1].Title = strings.Repeat("x", 256)
			case "invalidTestCase":
				problems[1].TestCases[0].Output = strings.Repeat("x", 65536)
			case "invalidExample":
				problems[1].Examples[0].Output = strings.Repeat("x", 65536)
			}
			tables := []string{"offer", "quiz", "language_quiz", "problem", "test_case", "example"}
			before := map[string]int{}
			for _, table := range tables {
				before[table] = f.Count(t, "SELECT COUNT(*) FROM "+table)
			}
			err := f.Storage.RegisterOffer(f.Ctx, offerID, offer, quizID, quiz, problems)
			if name != "success" {
				require.Error(t, err)
				for _, table := range tables {
					require.Equal(t, before[table], f.Count(t, "SELECT COUNT(*) FROM "+table), "rollback must restore %s", table)
				}
				return
			}
			require.NoError(t, err)
			stored, err := f.Storage.SelectOffer(f.Ctx, offerID)
			require.NoError(t, err)
			require.Equal(t, offer.Title, stored.Title)
			require.Equal(t, offer.CompanyID, stored.CompanyID)
			require.Equal(t, offer.About, stored.About)
			require.Equal(t, offer.Requirements, stored.Requirements)
			require.Equal(t, offer.Benefits, stored.Benefits)
			require.Equal(t, offer.MinWage, stored.MinWage)
			require.Equal(t, offer.MaxWage, stored.MaxWage)
			require.EqualValues(t, 1, stored.Status)
			storedQuiz, err := f.Storage.SelectQuizByOffer(f.Ctx, offerID)
			require.NoError(t, err)
			require.Equal(t, quizID, storedQuiz.ID)
			require.Equal(t, quiz.Duration, storedQuiz.Duration)
			languages, err := f.Storage.SelectLanguages(f.Ctx, quizID)
			require.NoError(t, err)
			ids := []int32{}
			for _, language := range languages {
				ids = append(ids, language.ID)
			}
			require.ElementsMatch(t, quiz.Languages, ids)
			storedProblems, err := f.Storage.SelectFullProblems(f.Ctx, quizID)
			require.NoError(t, err)
			require.Len(t, storedProblems, 2)
			byTitle := map[string]shared.Problem{}
			for _, problem := range storedProblems {
				byTitle[problem.Title] = problem
			}
			for _, expected := range problems {
				got, ok := byTitle[expected.Title]
				require.True(t, ok)
				require.Equal(t, expected.Description, got.Description)
				require.Equal(t, expected.TimeLimit, got.TimeLimit)
				cases := [][2]string{}
				for _, c := range got.TestCases {
					require.Equal(t, got.ID, c.ProblemID)
					cases = append(cases, [2]string{c.Input, c.Output})
				}
				expectedCases := [][2]string{}
				for _, c := range expected.TestCases {
					expectedCases = append(expectedCases, [2]string{c.Input, c.Output})
				}
				require.ElementsMatch(t, expectedCases, cases)
				require.Len(t, got.Examples, 1)
				require.Equal(t, expected.Examples[0].Input, got.Examples[0].Input)
				require.Equal(t, expected.Examples[0].Output, got.Examples[0].Output)
				require.Equal(t, got.ID, got.Examples[0].ProblemID)
			}
		})
	}
}

func TestSelectApplications(t *testing.T) {
	f := testutil.New(t)
	empty, err := f.Storage.SelectApplications(f.Ctx, f.QuizID)
	require.NoError(t, err)
	require.Empty(t, empty)
	firstPart := f.Participate(t, f.UserID)
	secondUser := uuid.NewString()
	f.AddUser(t, secondUser)
	secondPart := f.Participate(t, secondUser)
	noSubmissions := uuid.NewString()
	f.AddUser(t, noSubmissions)
	f.Participate(t, noSubmissions)
	secondProblem := uuid.NewString()
	f.AddProblem(t, secondProblem, f.QuizID)
	secondCase := uuid.NewString()
	f.Exec(t, "INSERT INTO test_case (id,problem_id,input,output) VALUES (?,?,'x','y')", secondCase, secondProblem)
	older := f.Submit(t, firstPart, f.ProblemID, 54)
	newer := f.Submit(t, firstPart, f.ProblemID, 62)
	worse := f.Submit(t, firstPart, f.ProblemID, 54)
	f.Exec(t, "UPDATE submission SET created_at='2026-01-01 00:00:01' WHERE id=?", older)
	f.Exec(t, "UPDATE submission SET created_at='2026-01-01 00:00:02' WHERE id=?", newer)
	for _, id := range []string{older, newer} {
		require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, uuid.NewString(), "Accepted"), id, f.CaseIDs[0]))
	}
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, uuid.NewString(), "Wrong Answer"), worse, f.CaseIDs[0]))
	otherProblemSubmission := f.Submit(t, firstPart, secondProblem, 62)
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, uuid.NewString(), "Accepted"), otherProblemSubmission, secondCase))
	secondSubmission := f.Submit(t, secondPart, f.ProblemID, 62)
	for _, id := range f.CaseIDs {
		require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, uuid.NewString(), "Accepted"), secondSubmission, id))
	}
	// A participation in a different quiz must not appear in this review.
	foreignUser := uuid.NewString()
	f.AddUser(t, foreignUser)
	foreignQuiz := uuid.NewString()
	f.Exec(t, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,15,?)", foreignQuiz, f.OfferID)
	require.NoError(t, f.Storage.Participate(f.Ctx, foreignUser, foreignQuiz))
	applications, err := f.Storage.SelectApplications(f.Ctx, f.QuizID)
	require.NoError(t, err)
	require.Len(t, applications, 3)
	byUser := map[string]shared.Application{}
	for _, a := range applications {
		byUser[a.Applicant.ID] = a
	}
	require.Empty(t, byUser[noSubmissions].Summary)
	first, ok := byUser[f.UserID]
	require.True(t, ok)
	require.Equal(t, firstPart, first.Participation.ID)
	require.Len(t, first.Summary, 2)
	byProblem := map[string]shared.Summary{}
	for _, s := range first.Summary {
		byProblem[s.Submission.ProblemID] = s
	}
	summary := byProblem[f.ProblemID]
	require.Equal(t, older, summary.Submission.ID)
	require.Equal(t, "C++", summary.Submission.Language)
	require.Equal(t, shared.Score{AcceptedTestCases: 1, TotalTestCases: 2}, summary.Score)
	require.Len(t, summary.Results, 2)
	for _, r := range summary.Results {
		require.Equal(t, older, r.SubmissionID)
		require.Contains(t, f.CaseIDs, r.TestCaseID)
	}
	summary = byProblem[secondProblem]
	require.Equal(t, otherProblemSubmission, summary.Submission.ID)
	require.Equal(t, shared.Score{AcceptedTestCases: 1, TotalTestCases: 1}, summary.Score)
	require.Len(t, summary.Results, 1)
	require.Equal(t, secondCase, summary.Results[0].TestCaseID)
	second, ok := byUser[secondUser]
	require.True(t, ok)
	require.Len(t, second.Summary, 1)
	require.Equal(t, secondSubmission, second.Summary[0].Submission.ID)
	require.Equal(t, shared.Score{AcceptedTestCases: 2, TotalTestCases: 2}, second.Summary[0].Score)
	for _, r := range second.Summary[0].Results {
		require.Equal(t, secondSubmission, r.SubmissionID)
		require.Equal(t, "Accepted", r.Status)
	}
}

func TestSelectOfferByUserAndArchiveOffer(t *testing.T) {
	f := testutil.New(t)
	stranger := uuid.NewString()
	f.AddUser(t, stranger)
	offer, err := f.Storage.SelectOfferByUser(f.Ctx, f.OfferID, f.UserID)
	require.NoError(t, err)
	require.Equal(t, f.OfferID, offer.ID)
	_, err = f.Storage.SelectOfferByUser(f.Ctx, f.OfferID, stranger)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, f.Storage.ArchiveOffer(f.Ctx, f.OfferID, stranger))
	require.Equal(t, 1, f.Count(t, "SELECT status FROM offer WHERE id=?", f.OfferID))
	require.NoError(t, f.Storage.ArchiveOffer(f.Ctx, f.OfferID, f.UserID))
	require.Equal(t, -1, f.Count(t, "SELECT status FROM offer WHERE id=?", f.OfferID))
	public, err := f.Storage.GetOffers(f.Ctx, 1)
	require.NoError(t, err)
	require.Empty(t, public)
	private, err := f.Storage.GetOffersByUser(f.Ctx, f.UserID, 1)
	require.NoError(t, err)
	require.Len(t, private, 1)
	require.Equal(t, f.OfferID, private[0].ID)
}

func TestSelectApplicationsWithoutSubmissions(t *testing.T) {
	f := testutil.New(t)
	f.Participate(t, f.UserID)
	applications, err := f.Storage.SelectApplications(f.Ctx, f.QuizID)
	require.NoError(t, err)
	require.Len(t, applications, 1)
	require.Empty(t, applications[0].Summary)
}
