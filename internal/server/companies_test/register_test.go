package companiestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/companies"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type registerStorage struct {
	mock.Mock
}

func (s *registerStorage) RegisterCompany(
	ctx context.Context,
	id string,
	userID string,
	name string,
	description string,
	website string,
	imageURL string,
) error {
	args := s.Called(ctx, id, userID, name, description, website, imageURL)
	return args.Error(0)
}

func registerInputFn(
	cloudinaryService shared.CloudinaryService,
	r *http.Request,
) (companies.CompanyRegInput, companies.CompanyRegErrors, bool) {
	return companies.CompanyRegInput{}, companies.CompanyRegErrors{}, false
}

func TestRegisterHandlerBadAuth(t *testing.T) {
	storage := new(registerStorage)
	handler := companies.CreateRegisterHandler(storage, invalidAuthRepo{}, &cldMock{}, registerInputFn, &templates{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterHandlerBadInput(t *testing.T) {
	invalidInputFn := func(
		cloudinaryService shared.CloudinaryService,
		r *http.Request,
	) (companies.CompanyRegInput, companies.CompanyRegErrors, bool) {
		return companies.CompanyRegInput{}, companies.CompanyRegErrors{}, true
	}
	storage := new(registerStorage)
	handler := companies.CreateRegisterHandler(storage, authRepo{}, &cldMock{}, invalidInputFn, &invalidTemplates{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(
		cloudinaryService shared.CloudinaryService,
		r *http.Request,
	) (companies.CompanyRegInput, companies.CompanyRegErrors, bool) {
		return companies.CompanyRegInput{}, companies.CompanyRegErrors{}, true
	}
	storage := new(registerStorage)
	handler := companies.CreateRegisterHandler(storage, authRepo{}, &cldMock{}, invalidInputFn, &templates{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterHandlerBadStorageRegisterCompany(t *testing.T) {
	storage := new(registerStorage)
	storage.On(
		"RegisterCompany",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := companies.CreateRegisterHandler(storage, authRepo{}, &cldMock{}, registerInputFn, &templates{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterHandler(t *testing.T) {
	prefix := "/c/"
	storage := new(registerStorage)
	storage.On(
		"RegisterCompany",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := companies.CreateRegisterHandler(storage, authRepo{}, &cldMock{}, registerInputFn, &templates{}, prefix)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK{
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	redirectPath := w.Header().Get("HX-Redirect")
	if !strings.HasPrefix(redirectPath, prefix) {
		t.Errorf("expected prefix %s in %s", prefix, redirectPath)
	}
}
