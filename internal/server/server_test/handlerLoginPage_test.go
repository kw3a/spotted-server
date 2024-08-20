package servertest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
)

func TestLoginPageHandlerBadTemplate(t *testing.T) {
	handler := server.CreateLoginPageHandler(&invalidTemplates{})
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
	handler := server.CreateLoginPageHandler(&templates{})
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
