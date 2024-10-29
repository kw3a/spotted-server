package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"golang.org/x/crypto/bcrypt"
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

func (mysql *MysqlStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error) {
	participation, err := mysql.Queries.ParticipationStatus(ctx, database.ParticipationStatusParams{
		UserID: userID,
		QuizID: quizID,
	})
	if err != nil {
		return ParticipationData{}, err
	}
	relativeTime := RelativeTime(participation.ExpiresAt)
	return ParticipationData{
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

func (mysql *MysqlStorage) SelectQuiz(ctx context.Context, quizID string) (Offer, error) {
	dbQuiz, err := mysql.Queries.GetQuiz(ctx, quizID)
	if err != nil {
		return Offer{}, err
	}
	languages, err := ConvertLanguages(dbQuiz.Languages)
	if err != nil {
		return Offer{}, err
	}
	relativeTime:= RelativeTime(dbQuiz.CreatedAt)
	return Offer{
		ID:           dbQuiz.ID,
		Title:        dbQuiz.Title,
		Description:  dbQuiz.Description,
		Duration:     dbQuiz.Duration,
		Author:       dbQuiz.Author,
		MinWage:      dbQuiz.MinWage,
		MaxWage:      dbQuiz.MaxWage,
		RelativeTime: relativeTime,
		Languages:    languages,
	}, nil
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

func (mysql *MysqlStorage) SelectOffers(ctx context.Context) ([]PartialOffer, error) {
	quizzes, err := mysql.Queries.GetQuizzes(ctx)
	if err != nil {
		return nil, err
	}
	offers := []PartialOffer{}
	for _, quiz := range quizzes {
		relativeTime := RelativeTime(quiz.CreatedAt)
		offers = append(offers, PartialOffer{
			QuizID:       quiz.ID,
			Title:        quiz.Title,
			Author:       quiz.Author,
			MinWage:      quiz.MinWage,
			MaxWage:      quiz.MaxWage,
			RelativeTime: relativeTime,
		})
	}
	return offers, nil
}

func RelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		duration = -duration
		if duration < time.Minute {
			return "in a few seconds"
		} else if duration < time.Hour {
			return fmt.Sprintf("in %d minutes", int(duration.Minutes()))
		} else if duration < 24*time.Hour {
			return fmt.Sprintf("in %d hours", int(duration.Hours()))
		} else {
			days := int(duration.Hours() / 24)
			return fmt.Sprintf("in %d days", days)
		}
	}
	if duration < time.Minute {
		return "justo ahora" 
	} else if duration < time.Hour {
		return fmt.Sprintf("hace %d minutos", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("hace %d horas", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("hace %d días", days)
	}
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
func (s MysqlStorage) EndQuiz(ctx context.Context, userID, quizID string) error {
	return s.Queries.EndParticipation(ctx, database.EndParticipationParams{
		ExpiresAt: time.Now(),
		UserID:    userID,
		QuizID:    quizID,
	})

}

func CheckPasswordHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func (s *MysqlStorage) GetUserID(ctx context.Context, email, password string) (string, error) {
	dbUser, err := s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	err = CheckPasswordHash(dbUser.Password, password)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	return dbUser.ID, nil
}
func (s *MysqlStorage) IsRegistered(ctx context.Context, refreshToken string) error {
	tokenExists, err := s.Queries.VerifyRefreshJWT(ctx, refreshToken)
	if err != nil {
		return err
	}
	if !tokenExists {
		return errors.New("token does not exist")
	}
	return nil
}
func (s *MysqlStorage) Save(ctx context.Context, refreshToken string) error {
	return s.Queries.SaveRefreshJWT(ctx, database.SaveRefreshJWTParams{
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	})
}
func (s *MysqlStorage) Revoke(ctx context.Context, refreshToken string) error {
	err := s.IsRegistered(ctx, refreshToken)
	if err != nil {
		return err
	}
	return s.Queries.RevokeRefreshJWT(ctx, refreshToken)
}

func (s *MysqlStorage) GetUser(ctx context.Context, userID string) (database.User, error) {
	return s.Queries.GetUser(ctx, userID)
}
