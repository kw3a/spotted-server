package offerstest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type adminStorage struct {
	mock.Mock
}

func (s *adminStorage) SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error) {
	args := s.Called(ctx, params)
	return args.Get(0).([]shared.Offer), args.Error(1)
}


func TestAdminHandlerBadAuth(t *testing.T) {
	storage := new(adminStorage)
	handler := offers.CreateOffersAdminHandler(invalidAuthRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAdminHandlerVisitor(t *testing.T) {
	storage := new(adminStorage)
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	handler := offers.CreateOffersAdminHandler(authz, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAdminHandlerBadStorage(t *testing.T) {
	storage := new(adminStorage)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, errors.New("error"))
	handler := offers.CreateOffersAdminHandler(authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAdminHandlerBadTemplate(t *testing.T) {
	storage := new(adminStorage)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{}, nil)
	handler := offers.CreateOffersAdminHandler(authRepo{}, storage, &invalidTemplates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAdminHandler(t *testing.T) {
	storage := new(adminStorage)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{}, nil)
	handler := offers.CreateOffersAdminHandler(authRepo{}, storage, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
