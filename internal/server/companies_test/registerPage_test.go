package companiestest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/companies"
	"github.com/stretchr/testify/mock"
)

func TestRegisterPageHandlerBadAuth(t *testing.T) {
	handler := companies.CreateRegisterPageHandler(&templates{}, invalidAuthRepo{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterPageHandlerVisitor(t *testing.T) {
	path := "/"
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	handler := companies.CreateRegisterPageHandler(&templates{}, authz, path)
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

func TestRegisterPageHandlerBadTemplate(t *testing.T) {
	handler := companies.CreateRegisterPageHandler(&invalidTemplates{}, &authRepo{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterPageHandler(t *testing.T) {
	handler := companies.CreateRegisterPageHandler(&templates{}, &authRepo{}, "")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
