package inc1test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/db_integration_test/testutil"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/require"
)

func TestCreateSubmission(t *testing.T) {
	f := testutil.New(t)
	otherProblem := uuid.NewString()
	f.AddProblem(t, otherProblem, f.QuizID)
	f.Exec(t, "INSERT INTO test_case (id,problem_id,input,output) VALUES (?,?,'other','other')", uuid.NewString(), otherProblem)
	participation := f.Participate(t, f.UserID)
	submission := f.Submit(t, participation, f.ProblemID, 54)
	results, err := f.Storage.GetResults(f.Ctx, f.ProblemID, submission)
	require.NoError(t, err)
	require.Len(t, results, 2)
	ids := []string{}
	for _, result := range results {
		ids = append(ids, result.TestCaseID)
		require.Empty(t, result.Status)
		require.Zero(t, result.Time)
	}
	require.ElementsMatch(t, f.CaseIDs, ids)
	results, err = f.Storage.GetResults(f.Ctx, otherProblem, submission)
	require.NoError(t, err)
	require.Empty(t, results)
	invalid := uuid.NewString()
	require.Error(t, f.Storage.CreateSubmission(f.Ctx, invalid, participation, f.ProblemID, "source", 999))
	require.Zero(t, f.Count(t, "SELECT COUNT(*) FROM submission WHERE id=?", invalid))
	require.Zero(t, f.Count(t, "SELECT COUNT(*) FROM test_case_result WHERE submission_id=?", invalid))
}

func TestUpdateTestCaseResult(t *testing.T) {
	f := testutil.New(t)
	part := f.Participate(t, f.UserID)
	submission := f.Submit(t, part, f.ProblemID, 54)
	other := f.Submit(t, part, f.ProblemID, 54)
	token := uuid.NewString()
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, token, "Accepted"), submission, f.CaseIDs[0]))
	var id, output, status string
	var elapsed, memory int
	require.NoError(t, f.DB.QueryRowContext(f.Ctx, "SELECT id,output,status,time,memory FROM test_case_result WHERE submission_id=? AND test_case_id=?", submission, f.CaseIDs[0]).Scan(&id, &output, &status, &elapsed, &memory))
	require.Equal(t, token, id)
	require.Equal(t, "output\n", output)
	require.Equal(t, "Accepted", status)
	require.Equal(t, 123, elapsed)
	require.Equal(t, 1024, memory)
	require.Equal(t, 1, f.Count(t, "SELECT accepted_test_cases FROM submission WHERE id=?", submission))
	require.Equal(t, 1, f.Count(t, "SELECT COUNT(*) FROM test_case_result WHERE submission_id=? AND status=''", submission))
	require.Equal(t, 2, f.Count(t, "SELECT COUNT(*) FROM test_case_result WHERE submission_id=? AND status=''", other))
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, testutil.Callback(t, uuid.NewString(), "Wrong Answer"), submission, f.CaseIDs[1]))
	require.Equal(t, 1, f.Count(t, "SELECT accepted_test_cases FROM submission WHERE id=?", submission))
}

func TestUpdateTestCaseResultRepeatedAccepted(t *testing.T) {
	f := testutil.New(t)
	part := f.Participate(t, f.UserID)
	submission := f.Submit(t, part, f.ProblemID, 54)
	input := testutil.Callback(t, uuid.NewString(), "Accepted")
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, input, submission, f.CaseIDs[0]))
	require.NoError(t, f.Storage.UpdateTestCaseResult(f.Ctx, input, submission, f.CaseIDs[0]))
	require.Equal(t, 1, f.Count(t, "SELECT accepted_test_cases FROM submission WHERE id=?", submission), "repeated callback must not count the same test case twice")
}

func TestBestSubmissionAndScore(t *testing.T) {
	f := testutil.New(t)
	part := f.Participate(t, f.UserID)
	score, err := f.Storage.SelectScore(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Equal(t, shared.Score{TotalTestCases: 2}, score)
	best, err := f.Storage.BestSubmission(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Empty(t, best.ID)
	older := f.Submit(t, part, f.ProblemID, 54)
	newer := f.Submit(t, part, f.ProblemID, 62)
	f.Exec(t, "UPDATE submission SET accepted_test_cases=1,created_at='2026-01-01 00:00:01' WHERE id=?", older)
	f.Exec(t, "UPDATE submission SET accepted_test_cases=1,created_at='2026-01-01 00:00:02' WHERE id=?", newer)
	otherUser := uuid.NewString()
	f.AddUser(t, otherUser)
	otherPart := f.Participate(t, otherUser)
	other := f.Submit(t, otherPart, f.ProblemID, 54)
	f.Exec(t, "UPDATE submission SET accepted_test_cases=2 WHERE id=?", other)
	otherProblem := uuid.NewString()
	f.AddProblem(t, otherProblem, f.QuizID)
	unrelated := f.Submit(t, part, otherProblem, 54)
	f.Exec(t, "UPDATE submission SET accepted_test_cases=2 WHERE id=?", unrelated)
	best, err = f.Storage.BestSubmission(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Equal(t, older, best.ID)
	require.Equal(t, "C++", best.Language)
	score, err = f.Storage.SelectScore(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Equal(t, shared.Score{AcceptedTestCases: 1, TotalTestCases: 2}, score)
	f.Exec(t, "UPDATE submission SET accepted_test_cases=2 WHERE id=?", newer)
	best, err = f.Storage.BestSubmission(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Equal(t, newer, best.ID)
	score, err = f.Storage.SelectScore(f.Ctx, f.UserID, f.ProblemID)
	require.NoError(t, err)
	require.Equal(t, 2, score.AcceptedTestCases)
}

func TestParticipationAndEndQuiz(t *testing.T) {
	f := testutil.New(t)
	_, err := f.Storage.ParticipationStatus(f.Ctx, f.UserID, f.QuizID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	f.Participate(t, f.UserID)
	p, err := f.Storage.ParticipationStatus(f.Ctx, f.UserID, f.QuizID)
	require.NoError(t, err)
	require.WithinDuration(t, p.CreatedAt.Add(30*time.Minute), p.ExpiresAt, time.Second)
	require.Error(t, f.Storage.Participate(f.Ctx, f.UserID, f.QuizID))
	require.Equal(t, 1, f.Count(t, "SELECT COUNT(*) FROM participation WHERE user_id=? AND quiz_id=?", f.UserID, f.QuizID))
	otherUser := uuid.NewString()
	f.AddUser(t, otherUser)
	f.Participate(t, otherUser)
	otherQuiz := uuid.NewString()
	f.Exec(t, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,45,?)", otherQuiz, f.OfferID)
	require.NoError(t, f.Storage.Participate(f.Ctx, f.UserID, otherQuiz))
	otherBefore, err := f.Storage.ParticipationStatus(f.Ctx, otherUser, f.QuizID)
	require.NoError(t, err)
	quizBefore, err := f.Storage.ParticipationStatus(f.Ctx, f.UserID, otherQuiz)
	require.NoError(t, err)
	var databaseNow time.Time
	require.NoError(t, f.DB.QueryRowContext(f.Ctx, "SELECT NOW()").Scan(&databaseNow))
	offer, err := f.Storage.EndQuiz(f.Ctx, f.UserID, f.QuizID)
	require.NoError(t, err)
	require.Equal(t, f.OfferID, offer.ID)
	ended, err := f.Storage.ParticipationStatus(f.Ctx, f.UserID, f.QuizID)
	require.NoError(t, err)
	require.WithinDuration(t, databaseNow, ended.ExpiresAt, 3*time.Second)
	otherAfter, err := f.Storage.ParticipationStatus(f.Ctx, otherUser, f.QuizID)
	require.NoError(t, err)
	require.Equal(t, otherBefore.ExpiresAt, otherAfter.ExpiresAt)
	quizAfter, err := f.Storage.ParticipationStatus(f.Ctx, f.UserID, otherQuiz)
	require.NoError(t, err)
	require.Equal(t, quizBefore.ExpiresAt, quizAfter.ExpiresAt)
}

func TestInsertQuiz(t *testing.T) {
	for _, tc := range []struct {
		name      string
		languages []int32
		invalid   bool
	}{
		{"success", []int32{54, 62}, false}, {"invalidLanguage", []int32{54, 999}, true}, {"duplicateLanguage", []int32{54, 54}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := testutil.New(t)
			id := uuid.NewString()
			err := f.Storage.InsertQuiz(f.Ctx, id, f.OfferID, tc.languages, 45)
			if tc.invalid {
				require.Error(t, err)
				require.Zero(t, f.Count(t, "SELECT COUNT(*) FROM quiz WHERE id=?", id))
				require.Zero(t, f.Count(t, "SELECT COUNT(*) FROM language_quiz WHERE quiz_id=?", id))
				return
			}
			require.NoError(t, err)
			require.Equal(t, 45, f.Count(t, "SELECT duration FROM quiz WHERE id=?", id))
			languages, err := f.Storage.SelectLanguages(f.Ctx, id)
			require.NoError(t, err)
			ids := []int32{}
			for _, language := range languages {
				ids = append(ids, language.ID)
			}
			require.ElementsMatch(t, tc.languages, ids)
		})
	}
}

func TestSelectFullProblems(t *testing.T) {
	f := testutil.New(t)
	empty := uuid.NewString()
	f.AddProblem(t, empty, f.QuizID)
	otherQuiz := uuid.NewString()
	f.Exec(t, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,15,?)", otherQuiz, f.OfferID)
	foreign := uuid.NewString()
	f.AddProblem(t, foreign, otherQuiz)
	for _, problem := range []string{f.ProblemID, foreign} {
		f.Exec(t, "INSERT INTO example (id,problem_id,input,output) VALUES (?,?,'example input','example output')", uuid.NewString(), problem)
	}
	problems, err := f.Storage.SelectFullProblems(f.Ctx, f.QuizID)
	require.NoError(t, err)
	require.Len(t, problems, 2)
	byID := map[string]shared.Problem{}
	for _, problem := range problems {
		byID[problem.ID] = problem
	}
	require.Len(t, byID[f.ProblemID].TestCases, 2)
	require.Len(t, byID[f.ProblemID].Examples, 1)
	for _, result := range byID[f.ProblemID].TestCases {
		require.Equal(t, f.ProblemID, result.ProblemID)
	}
	require.Equal(t, f.ProblemID, byID[f.ProblemID].Examples[0].ProblemID)
	require.Empty(t, byID[empty].Examples)
	require.Empty(t, byID[empty].TestCases)
	noProblems := uuid.NewString()
	f.Exec(t, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,15,?)", noProblems, f.OfferID)
	problems, err = f.Storage.SelectFullProblems(f.Ctx, noProblems)
	require.NoError(t, err)
	require.Empty(t, problems)
}

func TestLastSrc(t *testing.T) {
	f := testutil.New(t)
	part := f.Participate(t, f.UserID)
	example, err := f.Storage.LastSrc(f.Ctx, f.UserID, f.ProblemID, 54)
	require.NoError(t, err)
	require.NotEmpty(t, example)
	older := f.Submit(t, part, f.ProblemID, 54)
	newer := f.Submit(t, part, f.ProblemID, 54)
	f.Exec(t, "UPDATE submission SET created_at='2026-01-01 00:00:01' WHERE id=?", older)
	f.Exec(t, "UPDATE submission SET created_at='2026-01-01 00:00:02' WHERE id=?", newer)
	otherUser := uuid.NewString()
	f.AddUser(t, otherUser)
	otherPart := f.Participate(t, otherUser)
	foreign := f.Submit(t, otherPart, f.ProblemID, 54)
	java := f.Submit(t, part, f.ProblemID, 62)
	otherProblem := uuid.NewString()
	f.AddProblem(t, otherProblem, f.QuizID)
	unrelated := f.Submit(t, part, otherProblem, 54)
	for _, id := range []string{foreign, java, unrelated} {
		f.Exec(t, "UPDATE submission SET created_at='2026-01-01 00:00:03' WHERE id=?", id)
	}
	src, err := f.Storage.LastSrc(f.Ctx, f.UserID, f.ProblemID, 54)
	require.NoError(t, err)
	require.Equal(t, "source "+newer, src)
	src, err = f.Storage.LastSrc(f.Ctx, f.UserID, f.ProblemID, 62)
	require.NoError(t, err)
	require.Equal(t, "source "+java, src)
}
