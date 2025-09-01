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

type emailStorage struct {
	mock.Mock
}

func (s *emailStorage) UpdateEmail(ctx context.Context, userID, email string) error {
	args := s.Called(ctx, userID, email)
	return args.Error(0)
}

func emailInputFn(r *http.Request) (profiles.EmailUpdateInput, profiles.EmailUpdateErrors, bool) {
	return profiles.EmailUpdateInput{}, profiles.EmailUpdateErrors{}, false
}

func TestEmailHandlerBadAuth(t *testing.T) {
	storage := new(emailStorage)
	handler := profiles.CreateUpdateEmailHandler(&templates{}, invalidAuthRepo{}, storage, emailInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestEmailHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.EmailUpdateInput, profiles.EmailUpdateErrors, bool) {
		return profiles.EmailUpdateInput{}, profiles.EmailUpdateErrors{EmailError: "invalid email"}, true
	}
	storage := new(emailStorage)
	handler := profiles.CreateUpdateEmailHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEmailHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.EmailUpdateInput, profiles.EmailUpdateErrors, bool) {
		return profiles.EmailUpdateInput{}, profiles.EmailUpdateErrors{EmailError: "invalid email"}, true
	}
	storage := new(emailStorage)
	handler := profiles.CreateUpdateEmailHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEmailHandlerBadStorageT(t *testing.T) {
	storage := new(emailStorage)
	storage.On("UpdateEmail", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	handler := profiles.CreateUpdateEmailHandler(&invalidTemplates{}, &authRepo{}, storage, emailInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEmailHandlerBadStorage(t *testing.T) {
	storage := new(emailStorage)
	storage.On("UpdateEmail", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	handler := profiles.CreateUpdateEmailHandler(&templates{}, &authRepo{}, storage, emailInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEmailHandlerBadTemplate(t *testing.T) {
	storage := new(emailStorage)
	storage.On("UpdateEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateUpdateEmailHandler(&invalidTemplates{}, &authRepo{}, storage, emailInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEmailHandler(t *testing.T) {
	storage := new(emailStorage)
	storage.On("UpdateEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateUpdateEmailHandler(&templates{}, &authRepo{}, storage, emailInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
