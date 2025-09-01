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

func preambleInputFn(r *http.Request) (offers.PreambleInput, error) {
	return offers.PreambleInput{}, nil
}

type preambleStorage struct {
	mock.Mock
}

func (s *preambleStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error) {
	args := s.Called(ctx, userID, quizID)
	return args.Get(0).(shared.Participation), args.Error(1)
}

func (s *preambleStorage) SelectOffer(ctx context.Context, id string) (shared.Offer, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(shared.Offer), args.Error(1)
}

func (s *preambleStorage) SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error) {
	args := s.Called(ctx, offerID)
	return args.Get(0).(shared.Quiz), args.Error(1)
}

func (s *preambleStorage) SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error) {
	args := s.Called(ctx, quizID)
	return args.Get(0).([]shared.Language), args.Error(1)
}

func TestPreambleHandlerBadAuth(t *testing.T) {
	storage := new(preambleStorage)
	handler := offers.CreateParticipationHandler(&templates{}, storage, invalidAuthRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPreambleHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (offers.PreambleInput, error) {
		return offers.PreambleInput{}, errors.New("error")
	}
	storage := new(preambleStorage)
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectOffer(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{}, errors.New("error"))
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerOfferStatus0(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 0}, nil)
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectQuiz(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 1}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, errors.New("error"))
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerBadStorageSelectLanguages(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 1}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, errors.New("error"))
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPreambleHandlerBadStorageParticipationStatus(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 1}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(shared.Participation{}, errors.New("error"))
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestPreambleHandlerBadTemplate(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 1}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(shared.Participation{}, nil)
	handler := offers.CreateParticipationHandler(&invalidTemplates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPreambleHandler(t *testing.T) {
	storage := new(preambleStorage)
	storage.On("SelectOffer", mock.Anything, mock.Anything).Return(shared.Offer{Status: 1}, nil)
	storage.On("SelectQuizByOffer", mock.Anything, mock.Anything).Return(shared.Quiz{}, nil)
	storage.On("SelectLanguages", mock.Anything, mock.Anything).Return([]shared.Language{}, nil)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, mock.Anything).Return(shared.Participation{}, nil)
	handler := offers.CreateParticipationHandler(&templates{}, storage, authRepo{}, preambleInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
