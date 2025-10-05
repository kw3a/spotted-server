package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/shopspring/decimal"
)

func (mysql *MysqlStorage) SelectTestCases(ctx context.Context, problemID string) ([]shared.TestCase, error) {
	dbTC, err := mysql.Queries.SelectTestCases(ctx, problemID)
	if err != nil {
		return []shared.TestCase{}, nil
	}
	test_cases := []shared.TestCase{}
	for _, test_case := range dbTC {
		tc := shared.TestCase{
			ID:        test_case.ID,
			Input:     test_case.Input,
			Output:    test_case.Output,
			ProblemID: test_case.ProblemID,
		}
		test_cases = append(test_cases, tc)
	}
	return test_cases, nil
}

func (mysql *MysqlStorage) GetResults(ctx context.Context, problemID, submissionID string) ([]shared.TestCaseResult, error) {
	dbTCResults, err := mysql.Queries.GetResults(ctx, database.GetResultsParams{
		ProblemID: problemID,
		ID:        submissionID,
	})
	if err != nil {
		return []shared.TestCaseResult{}, err
	}
	tcResults := []shared.TestCaseResult{}
	for _, dbExecTC := range dbTCResults {
		tcResults = append(tcResults, shared.TestCaseResult{
			Output:     dbExecTC.Output,
			Status:     dbExecTC.Status,
			Time:       dbExecTC.Time,
			Memory:     dbExecTC.Memory,
			TestCaseID: dbExecTC.TestCaseID,
		},
		)
	}
	return tcResults, nil
}

func (mysql *MysqlStorage) BestSubmission(ctx context.Context, applicantID string, problemID string) (shared.Submission, error) {
	best, err := mysql.Queries.BestSubmission(ctx, database.BestSubmissionParams{
		ProblemID: problemID,
		UserID:    applicantID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return shared.Submission{}, nil
		}
		return shared.Submission{}, err
	}
	return shared.Submission{
		ID:                best.ID,
		Src:               best.Src,
		AcceptedTestCases: int(best.AcceptedTestCases),
		ParticipationID:   best.ParticipationID,
		LanguageID:        best.LanguageID,
		ProblemID:         best.ProblemID,
		Language:          best.Language,
	}, nil
}

func (mysql *MysqlStorage) InsertQuiz(ctx context.Context, quizID string, offerID string, languages []int32, duration int32) error {
	tx, err := mysql.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := mysql.Queries.WithTx(tx)
	err = qtx.InsertQuiz(ctx, database.InsertQuizParams{
		ID:       quizID,
		OfferID:  offerID,
		Duration: duration,
	})
	if err != nil {
		fmt.Println(offerID)
		return err
	}
	for _, lang := range languages {
		err = qtx.InsertLanguageQuiz(ctx, database.InsertLanguageQuizParams{
			ID:         uuid.New().String(),
			QuizID:     quizID,
			LanguageID: lang,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (mysql *MysqlStorage) SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error) {
	dbQuiz, err := mysql.Queries.GetQuizByOffer(ctx, offerID)
	if err != nil {
		return shared.Quiz{}, err
	}
	return shared.Quiz{
		ID:       dbQuiz.ID,
		Duration: dbQuiz.Duration,
	}, nil
}

func (mysql *MysqlStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error) {
	participation, err := mysql.Queries.ParticipationStatus(ctx, database.ParticipationStatusParams{
		UserID: userID,
		QuizID: quizID,
	})
	if err != nil {
		return shared.Participation{}, err
	}
	relativeTime := RelativeTime(participation.ExpiresAt)
	return shared.Participation{
		ID:           participation.ID,
		CreatedAt:    participation.CreatedAt.Time,
		ExpiresAt:    participation.ExpiresAt,
		RelativeTime: relativeTime,
	}, nil
}

func (mysql *MysqlStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return mysql.Queries.Participate(ctx, database.ParticipateParams{
		ID:     uuid.New().String(),
		UserID: userID,
		QuizID: quizID,
	})
}

func (mysql *MysqlStorage) SelectProblemIDs(ctx context.Context, QuizID string) ([]string, error) {
	IDs, err := mysql.Queries.SelectProblemIDs(ctx, QuizID)
	if err != nil {
		return nil, err
	}
	if len(IDs) == 0 {
		return nil, errors.New("zero results")
	}
	return IDs, nil
}

func ConvertLanguages(languages sql.NullString) ([]string, error) {
	if !languages.Valid {
		return nil, errors.New("languages is NULL")
	}
	result := strings.Split(languages.String, ",")
	if len(result) == 0 {
		return nil, errors.New("no languages found")
	}
	return result, nil
}

func (mysql *MysqlStorage) SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error) {
	totalTestCases, err := mysql.Queries.TotalTestCases(ctx, problemID)
	if err != nil {
		return shared.Score{}, err
	}
	best, err := mysql.Queries.BestSubmission(ctx, database.BestSubmissionParams{
		ProblemID: problemID,
		UserID:    userID,
	})
	acceptedTestCases := int(best.AcceptedTestCases)
	if err == sql.ErrNoRows {
		acceptedTestCases = 0
	}
	return shared.Score{
		AcceptedTestCases: acceptedTestCases,
		TotalTestCases:    int(totalTestCases),
	}, nil
}

func (mysql *MysqlStorage) SelectProblems(ctx context.Context, quizID string) ([]shared.Problem, error) {
	problems, err := mysql.Queries.SelectProblems(ctx, quizID)
	if err != nil {
		return []shared.Problem{}, err
	}
	res := []shared.Problem{}
	for _, problem := range problems {
		p := shared.Problem{
			ID:          problem.ID,
			Title:       problem.Title,
			Description: problem.Description,
			MemoryLimit: problem.MemoryLimit,
			TimeLimit:   problem.TimeLimit * 1000,
		}
		res = append(res, p)
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectProblem(ctx context.Context, problemID string) (shared.Problem, error) {
	dbProblem, err := mysql.Queries.SelectProblem(ctx, problemID)
	if err != nil {
		return shared.Problem{}, err
	}
	return shared.Problem{
		ID:          dbProblem.ID,
		Title:       dbProblem.Title,
		Description: dbProblem.Description,
		MemoryLimit: dbProblem.MemoryLimit,
		TimeLimit:   dbProblem.TimeLimit,
	}, nil
}

func (mysql *MysqlStorage) SelectExamples(ctx context.Context, problemID string) ([]shared.Example, error) {
	dbExamples, err := mysql.Queries.SelectExamples(ctx, problemID)
	if err != nil {
		return []shared.Example{}, err
	}
	res := []shared.Example{}
	for _, dbExample := range dbExamples {
		res = append(res, shared.Example{
			Input:  dbExample.Input,
			Output: dbExample.Output,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	dbLastSubmission, err := mysql.Queries.LastSubmission(ctx, database.LastSubmissionParams{
		ProblemID:  problemID,
		LanguageID: languageID,
		UserID:     userID,
	})
	if err == sql.ErrNoRows {
		return ExampleCode(languageID)
	}
	if err != nil {
		return "", err
	}
	return dbLastSubmission, nil
}

func (s MysqlStorage) CreateSubmission(
	ctx context.Context,
	submissionID, participationID, problemID, src string,
	languageID int32,
) error {
	return s.Queries.CreateSubmission(ctx, database.CreateSubmissionParams{
		ID:              submissionID,
		Src:             src,
		LanguageID:      languageID,
		ProblemID:       problemID,
		ParticipationID: participationID,
	})
}

func (s MysqlStorage) GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error) {
	dbTestCases, err := s.Queries.GetTestCases(ctx, problemID)
	if err != nil {
		return nil, err
	}
	return ToTC(dbTestCases)
}

func ToTC(dbTestCases []database.GetTestCasesRow) ([]codejudge.TestCase, error) {
	if len(dbTestCases) == 0 {
		return []codejudge.TestCase{}, errors.New("empty database test cases")
	}
	res := []codejudge.TestCase{}
	for _, testCase := range dbTestCases {
		fTimeLimit := float64(testCase.TimeLimit)
		current := codejudge.TestCase{
			ID:             testCase.ID,
			TimeLimit:      fTimeLimit / 1000,
			MemoryLimit:    testCase.MemoryLimit,
			Input:          testCase.Input,
			ExpectedOutput: testCase.Output,
		}
		res = append(res, current)
	}
	return res, nil
}

// This method must work only when TCResult is empty
func (s MysqlStorage) UpdateTestCaseResult(
	ctx context.Context,
	input shared.CallbackJsonInput,
	submissionID, testCaseID string,
) error {
	input.Time = input.Time.Mul(decimal.NewFromInt32(1000))
	intTime := input.Time.IntPart()
	return s.Queries.UpdateTestCaseResult(ctx, database.UpdateTestCaseResultParams{
		ID:           sql.NullString{String: input.Token, Valid: true},
		Output:       input.Stdout,
		Status:       input.Status.Description,
		Time:         int32(intTime),
		Memory:       input.Memory,
		SubmissionID: submissionID,
		TestCaseID:   testCaseID,
	})
}

func (s MysqlStorage) EndQuiz(ctx context.Context, userID, quizID string) (shared.Offer, error) {
	err := s.Queries.EndParticipation(ctx, database.EndParticipationParams{
		ExpiresAt: time.Now(),
		UserID:    userID,
		QuizID:    quizID,
	})
	if err != nil {
		return shared.Offer{}, err
	}
	offer, _ := s.Queries.GetOfferByQuiz(ctx, quizID)
	return shared.Offer{
		ID: offer.ID,
	}, nil
}
