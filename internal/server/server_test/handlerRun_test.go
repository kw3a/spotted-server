package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

func runInputFn(r *http.Request) (server.RunInput, error) {
	return server.RunInput{}, nil
}

type judgeService struct {}
func (j *judgeService) Send(dbTestCases []codejudge.TestCase, submissionID, src string, languageID int32) ([]string, error) {
	return nil, nil
}
type invalidJudgeService struct {}
func (i *invalidJudgeService) Send(dbTestCases []codejudge.TestCase, submissionID, src string, languageID int32) ([]string, error) {
	return nil, errors.New("error")
}

type streamService struct {}
func (s *streamService) Register(name string, tokens []string, duration time.Duration) error {
	return nil
}
type invalidStreamService struct {}
func (i *invalidStreamService) Register(name string, tokens []string, duration time.Duration) error {
	return errors.New("error")
}

type runStorage struct{}
func (r *runStorage) CreateSubmission(ctx context.Context, submissionID string, participationID string, problemID string, src string, languageID int32) error {
	return nil
}
func (r *runStorage) GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error) {
	return nil, nil
}
func (r *runStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	expiresAt := time.Now().Add(time.Hour)
	return server.ParticipationData{ExpiresAt: expiresAt}, nil
}

type invalidRunStorage struct{}
func (i *invalidRunStorage) CreateSubmission(ctx context.Context, submissionID string, participationID string, problemID string, src string, languageID int32) error {
	return errors.New("error")
}
func (i *invalidRunStorage) GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error) {
	return nil, errors.New("error")
}
func (i *invalidRunStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	expired := time.Now().Add(-time.Hour)
	return server.ParticipationData{ExpiresAt: expired}, errors.New("error")
}

func ValidFormValues() map[string][]string {
	quizID := uuid.NewString()
	problemID := uuid.NewString()
	src := "src"
	languageID := "60"
	formValues := map[string][]string{
		"quizID":     {quizID},
		"problemID":  {problemID},
		"src":        {src},
		"languageID": {languageID},
	}
	return formValues
}

func TestGetRunInputBadQuizID(t *testing.T) {
	formValues := ValidFormValues()
	formValues["quizID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetRunInputBadProblemID(t *testing.T) {
	formValues := ValidFormValues()
	formValues["problemID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetRunInputBadSrc(t *testing.T) {
	formValues := ValidFormValues()
	formValues["src"][0] = ""
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetRunInputBadLanguageID(t *testing.T) {
	formValues := ValidFormValues()
	formValues["languageID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetRunInput(t *testing.T) {
	formValues := ValidFormValues()
	req := formRequest("GET", "/", formValues)
	input, err := server.GetRunInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != formValues["quizID"][0] {
		t.Error("invalid quiz ID")
	}
	if input.ProblemID != formValues["problemID"][0] {
		t.Error("invalid problem ID")
	}
	if input.Src != formValues["src"][0] {
		t.Error("invalid src")
	}
	intInputLanguageID := int(input.LanguageID)
	strInputLanguageID := strconv.Itoa(intInputLanguageID)
	if strInputLanguageID != formValues["languageID"][0] {
		t.Error("invalid language ID")
	}
}

func TestRunHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.RunInput, error) {
		return server.RunInput{}, errors.New("error")
	}
	handler := server.CreateRunHandler(
		&templates{},
		&runStorage{},
		&authRepo{},
		&streamService{},
		&judgeService{},
		60*time.Second,
		invalidInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadStorage(t *testing.T) {
	handler := server.CreateRunHandler(
		&templates{},
		&invalidRunStorage{},
		&authRepo{},
		&streamService{},
		&judgeService{},
		60*time.Second,
		runInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadAuth(t *testing.T) {
	handler := server.CreateRunHandler(
		&templates{},
		&runStorage{},
		&invalidAuthRepo{},
		&streamService{},
		&judgeService{},
		60*time.Second,
		runInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Error("expected status unauthorized")
	}
}

func TestRunHandlerBadStream(t *testing.T) {
	handler := server.CreateRunHandler(
		&templates{},
		&runStorage{},
		&authRepo{},
		&invalidStreamService{},
		&judgeService{},
		60*time.Second,
		runInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadJudge(t *testing.T) {
	handler := server.CreateRunHandler(
		&templates{},
		&runStorage{},
		&authRepo{},
		&streamService{},
		&invalidJudgeService{},
		60*time.Second,
		runInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadTemplate(t *testing.T) {
	handler := server.CreateRunHandler(
		&invalidTemplates{},
		&runStorage{},
		&authRepo{},
		&streamService{},
		&judgeService{},
		60*time.Second,
		runInputFn,
	)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected status internal server error")
	}
}
