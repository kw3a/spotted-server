package offerstest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/stretchr/testify/mock"
)

type archiveStorage struct {
	mock.Mock
}

func (s *archiveStorage) ArchiveOffer(ctx context.Context, offerID string, ownerID string) error {
	args := s.Called(ctx, offerID, ownerID)
	return args.Error(0)
}

func archiveInputFn(r *http.Request) (offers.OfferArchiveInput, error) {
	return offers.OfferArchiveInput{}, nil
}

func TestArchiveBadAuth(t *testing.T) {
	storage := new(archiveStorage)
	handler := offers.CreateArchiveHandler(archiveInputFn, invalidAuthRepo{}, storage)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestArchiveVisitor(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	storage := new(archiveStorage)
	handler := offers.CreateArchiveHandler(archiveInputFn, authz, storage)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestArchiveBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (offers.OfferArchiveInput, error) {
		return offers.OfferArchiveInput{}, errors.New("error")
	}
	storage := new(archiveStorage)
	handler := offers.CreateArchiveHandler(invalidInputFn, authRepo{}, storage)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestArchiveBadStorageArchiveOffer(t *testing.T) {
	storage := new(archiveStorage)
	storage.On("ArchiveOffer", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	handler := offers.CreateArchiveHandler(archiveInputFn, authRepo{}, storage)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestArchiveHandler(t *testing.T) {
	storage := new(archiveStorage)
	storage.On("ArchiveOffer", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := offers.CreateArchiveHandler(archiveInputFn, authRepo{}, storage)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
