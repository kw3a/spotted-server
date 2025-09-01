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

type listStorage struct {
	mock.Mock
}

func (s *listStorage) GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error) {
	args := s.Called(ctx, params)
	return args.Get(0).([]shared.Company), args.Error(1)
}

func listInputFn(r *http.Request) shared.CompanyQueryParams {
	return shared.CompanyQueryParams{}
}

func TestListHandlerBadAuth(t *testing.T) {
	storage := new(listStorage)
	handler := companies.CreateCompanyListPageHandler(invalidAuthRepo{}, &templates{}, storage, listInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestListHandlerBadStorageGetCompanies(t *testing.T) {
	storage := new(listStorage)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return([]shared.Company{}, fmt.Errorf("error"))
	handler := companies.CreateCompanyListPageHandler(authRepo{}, &templates{}, storage, listInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListHandlerBadTemplate(t *testing.T) {
	storage := new(listStorage)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return([]shared.Company{}, nil)
	handler := companies.CreateCompanyListPageHandler(authRepo{}, &invalidTemplates{}, storage, listInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestListHandler(t *testing.T) {
	storage := new(listStorage)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return([]shared.Company{}, nil)
	handler := companies.CreateCompanyListPageHandler(authRepo{}, &templates{}, storage, listInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
