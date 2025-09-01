package profilestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type descrStorage struct {
	mock.Mock
}

func (s *descrStorage) UpdateDescription(ctx context.Context, userID string, description string) error {
	args := s.Called(ctx, userID, description)
	return args.Error(0)
}

func descrInputFn(r *http.Request) (profiles.DescUpdateInput, profiles.DescUpdateErrors, bool) {
	return profiles.DescUpdateInput{}, profiles.DescUpdateErrors{}, false
}

func TestDescrHandlerBadAuth(t *testing.T) {
	storage := new(descrStorage)
	handler := profiles.CreateDescUpdateHandler(&templates{}, invalidAuthRepo{}, storage, descrInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDescrHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.DescUpdateInput, profiles.DescUpdateErrors, bool) {
		return profiles.DescUpdateInput{}, profiles.DescUpdateErrors{}, true
	}
	storage := new(descrStorage)
	handler := profiles.CreateDescUpdateHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDescrHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.DescUpdateInput, profiles.DescUpdateErrors, bool) {
		return profiles.DescUpdateInput{}, profiles.DescUpdateErrors{}, true
	}
	storage := new(descrStorage)
	handler := profiles.CreateDescUpdateHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDescrHandlerBadStorageT(t *testing.T) {
	storage := new(descrStorage)
	storage.On("UpdateDescription", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	handler := profiles.CreateDescUpdateHandler(&invalidTemplates{}, &authRepo{}, storage, descrInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDescrHandlerBadStorage(t *testing.T) {
	storage := new(descrStorage)
	storage.On("UpdateDescription", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	handler := profiles.CreateDescUpdateHandler(&templates{}, &authRepo{}, storage, descrInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDescrHandlerBadTemplate(t *testing.T) {
	storage := new(descrStorage)
	storage.On("UpdateDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateDescUpdateHandler(&invalidTemplates{}, &authRepo{}, storage, descrInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDescrHandler(t *testing.T) {
	storage := new(descrStorage)
	storage.On("UpdateDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateDescUpdateHandler(&templates{}, &authRepo{}, storage, descrInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
