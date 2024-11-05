package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
)

type jobPageStorage struct{}
func (j *jobPageStorage) SelectOffers(ctx context.Context) ([]server.PartialOffer, error) {
	return []server.PartialOffer{}, nil
}

type invalidJobPageStorage struct{}
func (i *invalidJobPageStorage) SelectOffers(ctx context.Context) ([]server.PartialOffer, error) {
	return nil, errors.New("error")
}

func TestJobPageHandlerBadAuth(t *testing.T) {
	handler := server.CreateJobOffersHandler(&invalidAuthRepo{}, &templates{}, &jobPageStorage{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Error("expected unauthorized")
	}
}

func TestJobPageHandlerBadStorage(t *testing.T) {
	handler := server.CreateJobOffersHandler(&authRepo{}, &templates{}, &invalidJobPageStorage{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected internal server error")
	}
}

func TestJobPageHandlerBadTemplates(t *testing.T) {
	handler := server.CreateJobOffersHandler(&authRepo{}, &invalidTemplates{}, &jobPageStorage{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected internal server error")
	}
}

func TestJobPageHandler(t *testing.T) {
	handler := server.CreateJobOffersHandler(&authRepo{}, &templates{}, &jobPageStorage{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
