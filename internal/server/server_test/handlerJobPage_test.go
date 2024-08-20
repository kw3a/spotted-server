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
func (j *jobPageStorage) SelectOffers(ctx context.Context) ([]server.Offer, error) {
	return []server.Offer{}, nil
}

type invalidJobPageStorage struct{}
func (i *invalidJobPageStorage) SelectOffers(ctx context.Context) ([]server.Offer, error) {
	return nil, errors.New("error")
}

func TestJobPageHandlerBadStorage(t *testing.T) {
	handler := server.CreateJobOffersHandler(&templates{}, &invalidJobPageStorage{})
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
	handler := server.CreateJobOffersHandler(&invalidTemplates{}, &jobPageStorage{})
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
