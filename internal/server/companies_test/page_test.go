package companiestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/companies"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type pageStorage struct {
	mock.Mock
}

func (s *pageStorage) GetCompanyByID(ctx context.Context, companyID string) (shared.Company, error) {
	args := s.Called(ctx, companyID)
	return args.Get(0).(shared.Company), args.Error(1)
}
func (s *pageStorage) SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error) {
	args := s.Called(ctx, params)
	return args.Get(0).([]shared.Offer), args.Error(1)
}

func pageInputFn(r *http.Request) (companies.CompanyPageInput, error) {
	return companies.CompanyPageInput{}, nil
}

func TestPageHandlerBadAuth(t *testing.T) {
	storage := new(pageStorage)
	handler := companies.CreateCompanyPageHandler(&templates{}, invalidAuthRepo{}, storage, pageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestPageHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (companies.CompanyPageInput, error) {
		return companies.CompanyPageInput{}, fmt.Errorf("error")
	}
	storage := new(pageStorage)
	handler := companies.CreateCompanyPageHandler(&templates{}, authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPageHandlerBadStorageGetCompany(t *testing.T) {
	storage := new(pageStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{}, fmt.Errorf("error"))
	handler := companies.CreateCompanyPageHandler(&templates{}, authRepo{}, storage, pageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPageHandlerBadStorageSelectOffers(t *testing.T) {
	storage := new(pageStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{}, nil)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, fmt.Errorf("error"))
	handler := companies.CreateCompanyPageHandler(&templates{}, authRepo{}, storage, pageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPageHandlerBadTemplate(t *testing.T) {
	storage := new(pageStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{}, nil)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	handler := companies.CreateCompanyPageHandler(&invalidTemplates{}, authRepo{}, storage, pageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPageHandler(t *testing.T) {
	storage := new(pageStorage)
	storage.On("GetCompanyByID", mock.Anything, mock.Anything).Return(shared.Company{}, nil)
	storage.On("SelectOffers", mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	handler := companies.CreateCompanyPageHandler(&templates{}, authRepo{}, storage, pageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
