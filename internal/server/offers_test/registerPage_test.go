package offerstest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

func TestRegisterPageHandlerBadAuth(t *testing.T) {
	storage := new(registerStorage)
	handler := offers.CreateRegisterPage(&invalidAuthRepo{}, &templates{}, storage, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterPageHandlerVisitor(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	storage := new(registerStorage)
	path := "/login"
	handler := offers.CreateRegisterPage(authz, &templates{}, storage, path)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Header().Get("Location")
	if location != path {
		t.Errorf("expected redirect to %q, got %q", path, location)
	}
}

func TestRegisterPageHandlerBadStorageGetCompanies(t *testing.T) {
	storage := new(registerStorage)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return([]shared.Company{}, fmt.Errorf("error"))
	path := "/login"
	handler := offers.CreateRegisterPage(&authRepo{}, &templates{}, storage, path)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Header().Get("Location")
	if location != path {
		t.Errorf("expected redirect to %q, got %q", path, location)
	}
}

func TestRegisterPageHandlerBadStorageGetLanguages(t *testing.T) {
	storage := new(registerStorage)
	c := make([]shared.Company, 2)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return(c, nil)
	storage.On("GetLanguages", mock.Anything).Return([]shared.Language{}, fmt.Errorf("error"))
	path := "/login"
	handler := offers.CreateRegisterPage(&authRepo{}, &templates{}, storage, path)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterPageHandlerBadTemplate(t *testing.T) {
	storage := new(registerStorage)
	c := make([]shared.Company, 2)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return(c, nil)
	storage.On("GetLanguages", mock.Anything).Return([]shared.Language{}, nil)
	path := "/login"
	handler := offers.CreateRegisterPage(&authRepo{}, &invalidTemplates{}, storage, path)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterPageHandler(t *testing.T) {
	storage := new(registerStorage)
	c := make([]shared.Company, 2)
	storage.On("GetCompanies", mock.Anything, mock.Anything).Return(c, nil)
	storage.On("GetLanguages", mock.Anything).Return([]shared.Language{}, nil)
	path := "/login"
	handler := offers.CreateRegisterPage(&authRepo{}, &templates{}, storage, path)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
