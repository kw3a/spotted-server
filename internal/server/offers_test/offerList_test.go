package offerstest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/offers"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type offerListStorage struct{}

func (j *offerListStorage) SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error) {
	return []shared.Offer{}, nil
}

type invalidOfferListStorage struct{}

func (i *invalidOfferListStorage) SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error) {
	return nil, errors.New("error")
}

func offerListFn(r *http.Request) (shared.OfferQueryParams, error) {
	return shared.OfferQueryParams{}, nil
}

func TestOfferListHandlerBadAuth(t *testing.T) {
	handler := offers.CreateOfferListHandler(offerListFn, &invalidAuthRepo{}, &offerListStorage{}, &templates{})
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

func TestOfferListHandlerBadInput(t *testing.T) {
	inputFn := func(r *http.Request) (shared.OfferQueryParams, error) {
		return shared.OfferQueryParams{}, fmt.Errorf("input error")
	}
	handler := offers.CreateOfferListHandler(inputFn, &invalidAuthRepo{}, &offerListStorage{}, &templates{})
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest{
		t.Errorf("expected: %v, got %v", http.StatusBadRequest, w.Code)
	}
}

func TestOfferListHandlerBadStorage(t *testing.T) {
	handler := offers.CreateOfferListHandler(offerListFn, &authRepo{}, &invalidOfferListStorage{}, &templates{})
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

func TestOfferListHandlerBadTemplates(t *testing.T) {
	handler := offers.CreateOfferListHandler(offerListFn, &authRepo{}, &offerListStorage{}, &invalidTemplates{})
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

func TestOfferListHandler(t *testing.T) {
	handler := offers.CreateOfferListHandler(offerListFn, &authRepo{}, &offerListStorage{}, &templates{})
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
