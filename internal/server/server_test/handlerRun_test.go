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
	"github.com/stretchr/testify/mock"
)

func runInputFn(r *http.Request) (server.RunInput, error) {
	return server.RunInput{}, nil
}

type judgeService struct{}

func (j *judgeService) Send(dbTestCases []codejudge.TestCase, submission codejudge.Submission) ([]string, error) {
	return nil, nil
}

type invalidJudgeService struct{}

func (i *invalidJudgeService) Send(dbTestCases []codejudge.TestCase, submission codejudge.Submission) ([]string, error) {
	return nil, errors.New("error")
}

type streamService struct{}

func (s *streamService) Register(name string, tokens []string, duration time.Duration) error {
	return nil
}

type invalidStreamService struct{}

func (i *invalidStreamService) Register(name string, tokens []string, duration time.Duration) error {
	return errors.New("error")
}

type runStorage struct {
	mock.Mock
}

func (r *runStorage) CreateSubmission(ctx context.Context, submissionID string, participationID string, problemID string, src string, languageID int32) error {
	args := r.Called(ctx, submissionID, participationID, problemID, src, languageID)
	return args.Error(0)
}
func (r *runStorage) GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error) {
	args := r.Called(ctx, problemID)
	return args.Get(0).([]codejudge.TestCase), args.Error(1)
}
func (r *runStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	args := r.Called(ctx, userID, quizID)
	return args.Get(0).(server.ParticipationData), args.Error(1)
}

func validRunFormValues() map[string][]string {
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
	formValues := validRunFormValues()
	formValues["quizID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRunInputBadProblemID(t *testing.T) {
	formValues := validRunFormValues()
	formValues["problemID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRunInputBadSrc(t *testing.T) {
	formValues := validRunFormValues()
	formValues["src"][0] = ""
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRunInputBadLanguageID(t *testing.T) {
	formValues := validRunFormValues()
	formValues["languageID"][0] = "invalid"
	req := formRequest("GET", "/", formValues)
	_, err := server.GetRunInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRunInput(t *testing.T) {
	formValues := validRunFormValues()
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

func TestRunHandlerBadStorageParticipationStatus(t *testing.T) {
	storage := new(runStorage)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, errors.New("error"))
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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

func TestRunHandlerBadStorageParticipationExpired(t *testing.T) {
	storage := new(runStorage)
	expired := server.ParticipationData{ExpiresAt: time.Now().Add(-time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(expired, nil)
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	if w.Code != http.StatusUnauthorized {
		t.Error("expected status unauthorized")
	}
}

func TestRunHandlerBadStorageGetTestCases(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, errors.New("error"))
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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

func TestRunHandlerBadStorageCreateSubmission(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	if w.Code != http.StatusInternalServerError {
		t.Error("expected status internal server error")
	}
}

func TestRunHandlerBadAuth(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	if w.Code != http.StatusInternalServerError {
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadJudge(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	if w.Code != http.StatusInternalServerError {
		t.Error("expected status bad request")
	}
}

func TestRunHandlerBadTemplate(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateRunHandler(
		&invalidTemplates{},
		storage,
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
	if w.Code != http.StatusInternalServerError {
		t.Error("expected status internal server error")
	}
}

func TestRunHandler(t *testing.T) {
	storage := new(runStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("GetTestCases", mock.Anything, mock.Anything).Return([]codejudge.TestCase{}, nil)
	storage.On("CreateSubmission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateRunHandler(
		&templates{},
		storage,
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
	if w.Code != http.StatusOK {
		t.Error("expected status ok")
	}
}
