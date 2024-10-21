package servertest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
)

func TestLoginPageHandlerBadAuth(t *testing.T) {
	handler := server.CreateLoginPageHandler(&invalidAuthRepo{}, &templates{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Error("expected unauthorized")
	}
}

func TestLoginPageHandlerBadTemplate(t *testing.T) {
	handler := server.CreateLoginPageHandler(&authRepo{}, &invalidTemplates{})
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected internal server error")
	}
}

func TestLoginPageHandler(t *testing.T) {
	handler := server.CreateLoginPageHandler(&authRepo{} ,&templates{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
