package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
	"github.com/stretchr/testify/mock"
)

type preambleStorage struct {
	mock.Mock
}

func (s *preambleStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (server.ParticipationData, error) {
	args := s.Called(ctx, userID, quizID)
	return args.Get(0).(server.ParticipationData), args.Error(1)
}

func (s *preambleStorage) SelectProblemIDs(ctx context.Context, QuizID string) ([]string, error) {
	args := s.Called(ctx, QuizID)
	return args.Get(0).([]string), args.Error(1)
}
func (s *preambleStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	args := s.Called(ctx, problemID)
	return args.Get(0).(server.ProblemContent), args.Error(1)
}
func (s *preambleStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	args := s.Called(ctx, userID, problemID)
	return args.Get(0).(server.ScoreData), args.Error(1)
}
func (s *preambleStorage) SelectQuiz(ctx context.Context, id string) (server.Offer, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(server.Offer), args.Error(1)
}

func TestPreambleHandlerBadAuth(t *testing.T) {
	storage := new(preambleStorage)
	handler := server.CreateParticipationHandler(&templates{}, storage, invalidAuthRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPreambleHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.QuizPageInput, error) {
		return server.QuizPageInput{}, errors.New("error")
	}
	storage := new(preambleStorage)
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectQuiz(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, errors.New("error"))
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerBadStorageParticipationStatus(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, errors.New("error"))
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectProblemIDs(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, nil)
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return([]string{}, errors.New("error"))
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectProblem(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, nil)
	ids := []string{"1", "2"}
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return(ids, nil)
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, errors.New("error"))
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectScore(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, nil)
	ids := []string{"1", "2"}
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return(ids, nil)
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, errors.New("error"))
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPreambleHandlerBadTemplate(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, nil)
	ids := []string{"1", "2"}
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return(ids, nil)
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil)
	handler := server.CreateParticipationHandler(&invalidTemplates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPreambleHandler(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectQuiz", mock.Anything, mock.Anything).Return(server.Offer{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(server.ParticipationData{}, nil)
	ids := []string{"1", "2"}
	storage.On("SelectProblemIDs", mock.Anything, mock.Anything).Return(ids, nil)
	storage.On("SelectProblem", mock.Anything, mock.Anything).Return(server.ProblemContent{}, nil)
	storage.On("SelectScore", mock.Anything, mock.Anything, mock.Anything).Return(server.ScoreData{}, nil)
	handler := server.CreateParticipationHandler(&templates{}, storage, authRepo{}, quizPageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
