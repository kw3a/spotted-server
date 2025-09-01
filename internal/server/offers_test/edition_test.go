package offerstest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/stretchr/testify/mock"
)

type editionStorage struct {
	mock.Mock
}

func (s *editionStorage) InsertQuiz(
	ctx context.Context,
	quizID string,
	offerID string,
	languages []int32,
	duration int32,
) error {
	args := s.Called(ctx, quizID, offerID, languages, duration)
	return args.Error(0)
}

func editionInputFn(r *http.Request) (offers.OfferEdition, error) {
	return offers.OfferEdition{}, nil
}

func TestEditionHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (offers.OfferEdition, error) {
		return offers.OfferEdition{}, fmt.Errorf("error")
	}
	storage := new(editionStorage)
	handler := offers.CreateOfferEdition(storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEditionHandlerBadStorageInsertQuiz(t *testing.T) {
	storage := new(editionStorage)
	storage.On(
		"InsertQuiz",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := offers.CreateOfferEdition(storage, editionInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEditionHandler(t *testing.T) {
	storage := new(editionStorage)
	storage.On(
		"InsertQuiz",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := offers.CreateOfferEdition(storage, editionInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
