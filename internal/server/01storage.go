package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

type MysqlStorage struct {
	Queries *database.Queries
}

func NewMysqlStorage(dbURL string) (*MysqlStorage, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	queries := database.New(db)
	return &MysqlStorage{Queries: queries}, nil
}

func (mysql *MysqlStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (string, bool, error) {
	participation, err := mysql.Queries.ParticipationStatus(ctx, database.ParticipationStatusParams{
		UserID: userID,
		QuizID: quizID,
	})
	if err != nil {
		return "", false, err
	}
	inHour := false
	if time.Now().Before(participation.ExpiresAt) {
		inHour = true
	}
	return participation.ID, inHour, nil
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

func (mysql *MysqlStorage) SelectOffers(ctx context.Context) ([]Offer, error) {
	quizzes, err := mysql.Queries.GetQuizzes(ctx)
	if err != nil {
		return nil, err
	}
	offers := []Offer{}
	for _, quiz := range quizzes {
		offers = append(offers, Offer{
			ID:          quiz.ID,
			Title:       quiz.Title,
			Description: quiz.Description,
		})
	}
	return offers, nil
}

func (mysql *MysqlStorage) SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error) {
	dbLanguages, err := mysql.Queries.SelectLanguages(ctx, quizID)
	if err != nil {
		return nil, err
	}
	res := []LanguageSelector{}
	for _, dbLang := range dbLanguages {
		lang := LanguageSelector{
			LanguageID:    dbLang.ID,
			DisplayedName: dbLang.Name + " v" + strconv.Itoa(int(dbLang.Version)),
			SimpleName:    dbLang.Name,
		}
		res = append(res, lang)
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error) {
  totalTestCases, err := mysql.Queries.TotalTestCases(ctx, problemID)
  if err != nil {
    return ScoreData{}, err
  }
  best, err := mysql.Queries.BestSubmission(ctx, database.BestSubmissionParams{
    ProblemID: problemID,
    UserID:    userID,
  })
  acceptedTestCases := int(best.AcceptedTestCases)
  if err == sql.ErrNoRows {
    acceptedTestCases = 0
  }
  return ScoreData{
    AcceptedTestCases: acceptedTestCases,
    TotalTestCases:    int(totalTestCases),
  }, nil
}

func (mysql *MysqlStorage) SelectProblem(ctx context.Context, problemID string) (ProblemContent, error) {
	dbProblem, err := mysql.Queries.SelectProblem(ctx, problemID)
	if err != nil {
		return ProblemContent{}, err
	}
	return ProblemContent{
		Title:       dbProblem.Title,
		Description: dbProblem.Description,
	}, nil
}

func (mysql *MysqlStorage) SelectExamples(ctx context.Context, problemID string) ([]Example, error) {
	dbExamples, err := mysql.Queries.SelectExamples(ctx, problemID)
	if err != nil {
		return []Example{}, err
	}
	res := []Example{}
	for _, dbExample := range dbExamples {
		res = append(res, Example{
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

func (s MysqlStorage) CreateSubmission(ctx context.Context, submissionID, participationID, problemID, src string, languageID int32) error {
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
		current := codejudge.TestCase{
			ID:             testCase.ID,
			TimeLimit:      testCase.TimeLimit,
			MemoryLimit:    testCase.MemoryLimit,
			Input:          testCase.Input,
			ExpectedOutput: testCase.Output,
		}
		res = append(res, current)
	}
	return res, nil
}

// This method must work only when TCResult is empty
func (s MysqlStorage) UpdateTestCaseResult(ctx context.Context, input CallbackInput, submissionID, testCaseID string) error {
	return s.Queries.UpdateTestCaseResult(ctx, database.UpdateTestCaseResultParams{
    ID:           sql.NullString{String: input.Token, Valid: true},
    Status:       sql.NullString{String: input.Status.Description, Valid: true},
    Time:         sql.NullString{String: input.Time.String(), Valid: true},
    Memory:       sql.NullInt32{Int32: input.Memory, Valid: true},
		SubmissionID: submissionID,
		TestCaseID:   testCaseID,
	})
}
