package profilestest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type linkStorage struct {
	mock.Mock
}

func (s *linkStorage) CountLinks(ctx context.Context, userID string) (int32, error) {
	args := s.Called(ctx, userID)
	return int32(args.Int(0)), args.Error(1)
}

func (s *linkStorage) RegisterLink(ctx context.Context, linkID, userID, url, name string) error {
	args := s.Called(ctx, linkID, userID, url, name)
	return args.Error(0)
}

func (s *linkStorage) DeleteLink(ctx context.Context, userID, linkID string) error {
	args := s.Called(ctx, userID, linkID)
	return args.Error(0)
}

func linkRegisterInputFn(r *http.Request) (profiles.LinkRegisterInput, profiles.LinkRegisterError, bool) {
	return profiles.LinkRegisterInput{}, profiles.LinkRegisterError{}, false
}

func linkDeleteInputFn(r *http.Request) (profiles.LinkDeleteInput, error) {
	return profiles.LinkDeleteInput{LinkID: "123e4567-e89b-12d3-a456-426614174000"}, nil
}

func TestRegisterLinkHandlerBadAuth(t *testing.T) {
	storage := new(linkStorage)
	handler := profiles.CreateRegisterLinkHandler(&templates{}, invalidAuthRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterLinkHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.LinkRegisterInput, profiles.LinkRegisterError, bool) {
		return profiles.LinkRegisterInput{}, profiles.LinkRegisterError{NameError: "name required"}, true
	}
	storage := new(linkStorage)
	handler := profiles.CreateRegisterLinkHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterLinkHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.LinkRegisterInput, profiles.LinkRegisterError, bool) {
		return profiles.LinkRegisterInput{}, profiles.LinkRegisterError{NameError: "name required"}, true
	}
	storage := new(linkStorage)
	handler := profiles.CreateRegisterLinkHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterLinkHandlerBadStorageCountT(t *testing.T) {
	storage := new(linkStorage)
	storage.On("CountLinks", mock.Anything, mock.Anything).Return(0, fmt.Errorf("count error"))
	handler := profiles.CreateRegisterLinkHandler(&invalidTemplates{}, &authRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterLinkHandlerBadStorageCount(t *testing.T) {
	storage := new(linkStorage)
	storage.On("CountLinks", mock.Anything, mock.Anything).Return(0, fmt.Errorf("count error"))
	handler := profiles.CreateRegisterLinkHandler(&templates{}, &authRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterLinkHandlerBadStorageRegT(t *testing.T) {
	storage := new(linkStorage)
	storage.On("CountLinks", mock.Anything, mock.Anything).Return(0, nil)
	storage.On("RegisterLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("storage error"))
	handler := profiles.CreateRegisterLinkHandler(&templates{}, &authRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterLinkHandlerBadTemplate(t *testing.T) {
	storage := new(linkStorage)
	storage.On("CountLinks", mock.Anything, mock.Anything).Return(0, nil)
	storage.On("RegisterLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	handler := profiles.CreateRegisterLinkHandler(&invalidTemplates{}, &authRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "link-added" {
		t.Errorf("expected redirect %s, got %s", "link-added", w.Header().Get("HX-Trigger"))
	}
}

func TestRegisterLinkHandler(t *testing.T) {
	storage := new(linkStorage)
	storage.On("CountLinks", mock.Anything, mock.Anything).Return(0, nil)
	storage.On("RegisterLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	handler := profiles.CreateRegisterLinkHandler(&templates{}, &authRepo{}, storage, linkRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "link-added" {
		t.Errorf("expected redirect %s, got %s", "link-added", w.Header().Get("HX-Trigger"))
	}
}

func TestDeleteLinkHandlerBadAuth(t *testing.T) {
	storage := new(linkStorage)
	handler := profiles.CreateDeleteLinkHandler(invalidAuthRepo{}, storage, linkDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteLinkHandlerBadInput(t *testing.T) {
	storage := new(linkStorage)
	invalidInputFn := func(r *http.Request) (profiles.LinkDeleteInput, error) {
		return profiles.LinkDeleteInput{}, errors.New("invalid input")
	}
	handler := profiles.CreateDeleteLinkHandler(&authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteLinkHandlerStorageError(t *testing.T) {
	storage := new(linkStorage)
	storage.On("DeleteLink", mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("storage error"))
	handler := profiles.CreateDeleteLinkHandler(&authRepo{}, storage, linkDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeleteLinkHandler(t *testing.T) {
	storage := new(linkStorage)
	storage.On("DeleteLink", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	handler := profiles.CreateDeleteLinkHandler(&authRepo{}, storage, linkDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
