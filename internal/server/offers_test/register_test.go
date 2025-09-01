package offerstest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type registerStorage struct {
	mock.Mock
}

func (s *registerStorage) GetCompanyByID(ctx context.Context, companyID string) (shared.Company, error) {
	args := s.Called(ctx, companyID)
	return args.Get(0).(shared.Company), args.Error(1)
}
func (s *registerStorage) RegisterOffer(
	ctx context.Context,
	offerID string,
	offer shared.Offer,
	quizID string,
	quiz shared.Quiz,
	problems []shared.Problem,
) error {
	args := s.Called(ctx, ctx, offerID, quizID, quiz, problems)
	return args.Error(0)
}
func (s *registerStorage) GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error) {
	args := s.Called(ctx, params)
	return args.Get(0).([]shared.Company), args.Error(1)
}
func (s *registerStorage) GetLanguages(ctx context.Context) ([]shared.Language, error) {
	args := s.Called(ctx)
	return args.Get(0).([]shared.Language), args.Error(1)
}

func registerInputFn(r *http.Request) (offers.OfferRegInput, error) {
	return offers.OfferRegInput{}, nil
}

func TestRegisterHandlerBadAuth(t *testing.T) {
	storage := new(registerStorage)
	handler := offers.CreateRegisterHandler(&templates{}, &invalidAuthRepo{}, storage, "", registerInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (offers.OfferRegInput, error) {
		return offers.OfferRegInput{}, fmt.Errorf("error")
	}
	storage := new(registerStorage)
	handler := offers.CreateRegisterHandler(&templates{}, &authRepo{}, storage, "", invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterHandlerBadStorageGetCompany(t *testing.T) {
	storage := new(registerStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{}, fmt.Errorf("error"))
	handler := offers.CreateRegisterHandler(&templates{}, &authRepo{}, storage, "", registerInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterHandlerNotOwner(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{ID: "1"}, nil)
	storage := new(registerStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{UserID: "2"}, nil)
	handler := offers.CreateRegisterHandler(&templates{}, authz, storage, "", registerInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterHandlerBadStorageRegisterOffer(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{ID: "1"}, nil)
	storage := new(registerStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{UserID: "1"}, nil)
	storage.On(
		"RegisterOffer",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := offers.CreateRegisterHandler(&templates{}, authz, storage, "", registerInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterHandler(t *testing.T) {
	prefix := "red/"
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{ID: "1"}, nil)
	storage := new(registerStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{UserID: "1"}, nil)
	storage.On(
		"RegisterOffer",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := offers.CreateRegisterHandler(&templates{}, authz, storage, prefix, registerInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	redirectPath := w.Header().Get("HX-Redirect")
	if !strings.HasPrefix(redirectPath, prefix) {
		t.Errorf("expected prefix %s in %s", prefix, redirectPath)
	}
}
