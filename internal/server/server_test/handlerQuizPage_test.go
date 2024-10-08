package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
	"github.com/stretchr/testify/mock"
)

func quizPageInputFn(r *http.Request) (server.QuizPageInput, error) {
	return server.QuizPageInput{}, nil
}
func selectProblemsFn(problems []string) string {
	return ""
}
func selectLanguagesFn(languages []server.LanguageSelector) server.LanguageSelector {
	return server.LanguageSelector{}
}
func enumerateFn([]string) []server.ProblemSelector{
	return []server.ProblemSelector{}
}
type quizPageStorage struct{
	mock.Mock
}

func (q *quizPageStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	args := q.Called(ctx, userID, problemID, languageID)
	return args.String(0), args.Error(1)
}
func (q *quizPageStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	args := q.Called(ctx, userID, quizID)
	return args.Get(0).(server.ParticipationData), args.Error(1)
}
func (q *quizPageStorage) SelectExamples(ctx context.Context, problemID string) ([]server.Example, error) {
	args := q.Called(ctx, problemID)
	return args.Get(0).([]server.Example), args.Error(1)
}
func (q *quizPageStorage) SelectLanguages(ctx context.Context, quizID string) ([]server.LanguageSelector, error) {
	args := q.Called(ctx, quizID)
	return args.Get(0).([]server.LanguageSelector), args.Error(1)
}
func (q *quizPageStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	args := q.Called(ctx, problemID)
	return args.Get(0).(server.ProblemContent), args.Error(1)
}
func (q *quizPageStorage) SelectProblemIDs(ctx context.Context, quizID string) ([]string, error) {
	args := q.Called(ctx, quizID)
	return args.Get(0).([]string), args.Error(1)
}
func (q *quizPageStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	args := q.Called(ctx, userID, problemID)
	return args.Get(0).(server.ScoreData), args.Error(1)
}

func TestGetQuizPageInputEmpty(t *testing.T) {
	params := map[string]string{
		"quizID": "",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	reqWithParams := WithUrlParams(req, params)
	_, err := server.GetQuizPageInput(reqWithParams)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetQuizPageInputBadQuizID(t *testing.T) {
	params := map[string]string{
		"quizID": "invalid",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	reqWithParams := WithUrlParams(req, params)
	_, err := server.GetQuizPageInput(reqWithParams)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetQuizPageInput(t *testing.T) {
	quizID := uuid.NewString()
	params := map[string]string{
		"quizID": quizID,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	reqWithParams := WithUrlParams(req, params)
	input, err := server.GetQuizPageInput(reqWithParams)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestQuizPageHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.QuizPageInput, error) {
		return server.QuizPageInput{}, errors.New("error")
	}
	handler := server.CreateQuizPageHandler(
		&templates{},
		&quizPageStorage{}, 
		&authRepo{}, 
		"/", 
		invalidInputFn, 
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestQuizPageHandlerBadAuth(t *testing.T) {
	handler := server.CreateQuizPageHandler(
		&templates{},
		&quizPageStorage{}, 
		&invalidAuthRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusSeeOther{
		t.Error("expected see other")
	}
	if w.Header().Get("Location") != "/" {
		t.Error("expected redirect to /")
	}
}

func TestQuizPageHandlerBadStorageParticipationStatus(t *testing.T) {
	storage := new(quizPageStorage)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestQuizPageHandlerBadStorageParticipationStatusExpired(t *testing.T) {
	storage := new(quizPageStorage)
	expired := server.ParticipationData{ExpiresAt: time.Now().Add(-time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(expired, nil)
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized{
		t.Error("expected unauthorized")
	}
}

func TestQuizPageHandlerBadStorageSelectProblemIDs(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestQuizPageHandlerBadStorageSelectScore(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandlerBadStorageSelectProblem(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandlerBadStorageSelectExamples(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil) 
	storage.On("SelectExamples", mock.Anything, mock.Anything).Return([]server.Example{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandlerBadStorageSelectLanguages(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil) 
	storage.On("SelectExamples", mock.Anything, mock.Anything).Return([]server.Example{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]server.LanguageSelector{}, errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandlerBadStorageLastSource(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil) 
	storage.On("SelectExamples", mock.Anything, mock.Anything).Return([]server.Example{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]server.LanguageSelector{}, nil) 
	storage.On("LastSrc", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("error"))
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandlerBadTemplate(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil) 
	storage.On("SelectExamples", mock.Anything, mock.Anything).Return([]server.Example{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]server.LanguageSelector{}, nil) 
	storage.On("LastSrc", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	handler := server.CreateQuizPageHandler(
		&invalidTemplates{},
		storage,
		&authRepo{},
		"/",
		quizPageInputFn,
		selectProblemsFn,
		selectLanguagesFn,
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError{
		t.Error("expected internal server error")
	}
}

func TestQuizPageHandler(t *testing.T) {
	storage := new(quizPageStorage)
	inTime := server.ParticipationData{ExpiresAt: time.Now().Add(time.Hour)}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(inTime, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil) 
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil) 
	storage.On("SelectExamples", mock.Anything, mock.Anything).Return([]server.Example{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]server.LanguageSelector{}, nil) 
	storage.On("LastSrc", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	handler := server.CreateQuizPageHandler(
		&templates{},
		storage,
		&authRepo{}, 
		"/", 
		quizPageInputFn,
		selectProblemsFn, 
		selectLanguagesFn, 
		enumerateFn,
	)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
