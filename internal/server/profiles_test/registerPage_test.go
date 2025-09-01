package profilestest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/profiles"
)

func TestRegisterPageHandlerBadAuth(t *testing.T) {
	handler := profiles.CreateRegPageHandler(invalidAuthRepo{}, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterPageHandlerBadTemplate(t *testing.T) {
	handler := profiles.CreateRegPageHandler(authRepo{}, &invalidTemplates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterPageHandler(t *testing.T) {
	handler := profiles.CreateRegPageHandler(authRepo{}, &templates{})
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
