package offerstest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type applicantsStorage struct {
	mock.Mock
}
func (s *applicantsStorage) SelectOfferByUser(ctx context.Context, id string, userID string) (shared.Offer, error) {
	args := s.Called(ctx, id, userID)
	return args.Get(0).(shared.Offer), args.Error(1)
}
func (s *applicantsStorage) SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error) {
	args := s.Called(ctx, offerID)
	return args.Get(0).(shared.Quiz), args.Error(1)
}
func (s *applicantsStorage) SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error) {
	args := s.Called(ctx, quizID)
	return args.Get(0).([]shared.Language), args.Error(1)
}
func (s *applicantsStorage) SelectApplications(ctx context.Context, quizID string) ([]shared.Application, error) {
	args := s.Called(ctx, quizID)
	return args.Get(0).([]shared.Application), args.Error(1)
}
func (s *applicantsStorage) SelectFullProblems(ctx context.Context, quizID string) ([]shared.Problem, error) {
	args := s.Called(ctx, quizID)
	return args.Get(0).([]shared.Problem), args.Error(1)
}

func applicantsInputFn(r *http.Request) (offers.ApplicantsInput, error) {
	return offers.ApplicantsInput{}, nil
}

func TestApplicantsHandlerBadAuth(t *testing.T) {
	storage := new(applicantsStorage)
	handler := offers.CreateApplicantsHandler(applicantsInputFn, invalidAuthRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestApplicantsHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (offers.ApplicantsInput, error) {
		return offers.ApplicantsInput{}, errors.New("error")
	}
	storage := new(applicantsStorage)
	handler := offers.CreateApplicantsHandler(invalidInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApplicantsHandlerBadStorageSelectOffer(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, errors.New("error"))
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApplicantsHandlerBadStorageSelectQuiz(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, errors.New("error"))
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApplicantsHandlerBadStorageSelectLanguages(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, errors.New("error"))
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApplicantsHandlerBadStorageSelectApplications(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("SelectApplications", mock.Anything, mock.Anything).Return([]shared.Application{}, errors.New("error"))
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApplicantsHandlerBadStorageSelectProblems(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("SelectApplications", mock.Anything, mock.Anything).Return([]shared.Application{}, nil)
	storage.On("SelectFullProblems", mock.Anything, mock.Anything).Return([]shared.Problem{}, errors.New("error"))
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApplicantsHandlerBadTemplate(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("SelectApplications", mock.Anything, mock.Anything).Return([]shared.Application{}, nil)
	storage.On("SelectFullProblems", mock.Anything, mock.Anything).Return([]shared.Problem{}, nil)
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &invalidTemplates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApplicantsHandler(t *testing.T) {
	storage := new(applicantsStorage)
	storage.On("SelectOfferByUser", mock.Anything, mock.Anything, mock.Anything).Return(shared.Offer{}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("SelectApplications", mock.Anything, mock.Anything).Return([]shared.Application{}, nil)
	storage.On("SelectFullProblems", mock.Anything, mock.Anything).Return([]shared.Problem{}, nil)
	handler := offers.CreateApplicantsHandler(applicantsInputFn, authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK{
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
