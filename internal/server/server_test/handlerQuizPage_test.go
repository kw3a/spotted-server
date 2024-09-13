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
type quizPageStorage struct{}

func (q *quizPageStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	return "", nil
}
func (q *quizPageStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	expiresAt := time.Now().Add(time.Hour)
	return server.ParticipationData{ExpiresAt: expiresAt}, nil
}
func (q *quizPageStorage) SelectExamples(ctx context.Context, problemID string) ([]server.Example, error) {
	return nil, nil
}
func (q *quizPageStorage) SelectLanguages(ctx context.Context, quizID string) ([]server.LanguageSelector, error) {
	return nil, nil
}
func (q *quizPageStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	return server.ProblemContent{}, nil
}
func (q *quizPageStorage) SelectProblemIDs(ctx context.Context, quizID string) ([]string, error) {
	return nil, nil
}
func (q *quizPageStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	return server.ScoreData{}, nil
}

type invalidQuizPageStorage struct{}

func (i *invalidQuizPageStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	return "", errors.New("error")
}
func (i *invalidQuizPageStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	expired := time.Now().Add(-time.Hour)
	return server.ParticipationData{ExpiresAt: expired}, errors.New("error")
}
func (i *invalidQuizPageStorage) SelectExamples(ctx context.Context, problemID string) ([]server.Example, error) {
	return nil, errors.New("error")
}
func (i *invalidQuizPageStorage) SelectLanguages(ctx context.Context, quizID string) ([]server.LanguageSelector, error) {
	return nil, errors.New("error")
}
func (i *invalidQuizPageStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	return server.ProblemContent{}, errors.New("error")
}
func (i *invalidQuizPageStorage) SelectProblemIDs(ctx context.Context, quizID string) ([]string, error) {
	return nil, errors.New("error")
}
func (i *invalidQuizPageStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	return server.ScoreData{}, errors.New("error")
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
}
func TestQuizPageHandlerBadStorage(t *testing.T) {
	handler := server.CreateQuizPageHandler(
		&templates{},
		&invalidQuizPageStorage{}, 
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
func TestQuizPageHandlerBadTemplate(t *testing.T) {
	handler := server.CreateQuizPageHandler(
		&invalidTemplates{},
		&quizPageStorage{},
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
	handler := server.CreateQuizPageHandler(
		&templates{},
		&quizPageStorage{}, 
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
