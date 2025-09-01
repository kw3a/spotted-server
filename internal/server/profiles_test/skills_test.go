package profilestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type skillStorage struct {
	mock.Mock
}

func (s *skillStorage) DeleteSkill(ctx context.Context, userID string, skillID string) error {
	args := s.Called(ctx, userID, skillID)
	return args.Error(0)
}
func (s *skillStorage) RegisterSkill(ctx context.Context, skillID string, userID string, name string) error {
	args := s.Called(ctx, skillID, userID, name)
	return args.Error(0)
}

func skillRegisterInputFn(r *http.Request) (profiles.SkillRegisterInput, profiles.SkillRegisterErrors, bool) {
	return profiles.SkillRegisterInput{}, profiles.SkillRegisterErrors{}, false
}

func skillDeleteInputFn(r *http.Request) (profiles.SkillDeleteInput, error) {
	return profiles.SkillDeleteInput{SkillID: "123e4567-e89b-12d3-a456-426614174000"}, nil
}

func TestRegisterSkillHandlerBadAuth(t *testing.T) {
	storage := new(skillStorage)
	handler := profiles.CreateRegisterSkillHandler(&templates{}, invalidAuthRepo{}, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterSkillHandlerVisitor(t *testing.T) {
	storage := new(skillStorage)
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	handler := profiles.CreateRegisterSkillHandler(&templates{}, authz, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterSkillHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.SkillRegisterInput, profiles.SkillRegisterErrors, bool) {
		return profiles.SkillRegisterInput{}, profiles.SkillRegisterErrors{}, true
	}
	storage := new(skillStorage)
	handler := profiles.CreateRegisterSkillHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterSkillHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.SkillRegisterInput, profiles.SkillRegisterErrors, bool) {
		return profiles.SkillRegisterInput{}, profiles.SkillRegisterErrors{}, true
	}
	storage := new(skillStorage)
	handler := profiles.CreateRegisterSkillHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterSkillHandlerBadStorageT(t *testing.T) {
	storage := new(skillStorage)
	storage.On("RegisterSkill", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	handler := profiles.CreateRegisterSkillHandler(&invalidTemplates{}, &authRepo{}, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRegisterSkillHandlerBadStorage(t *testing.T) {
	storage := new(skillStorage)
	storage.On("RegisterSkill", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	handler := profiles.CreateRegisterSkillHandler(&templates{}, &authRepo{}, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterSkillHandlerBadTemplate(t *testing.T) {
	storage := new(skillStorage)
	storage.On("RegisterSkill", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateRegisterSkillHandler(&invalidTemplates{}, &authRepo{}, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "skill-added" {
		t.Errorf("expected redirect %s, got %s", "skill-added", w.Header().Get("HX-Trigger"))
	}
}

func TestRegisterSkillHandler(t *testing.T) {
	storage := new(skillStorage)
	storage.On("RegisterSkill", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateRegisterSkillHandler(&templates{}, &authRepo{}, storage, skillRegisterInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "skill-added" {
		t.Errorf("expected redirect %s, got %s", "skill-added", w.Header().Get("HX-Trigger"))
	}
}

func TestDeleteSkillHandlerBadAuth(t *testing.T) {
	storage := new(skillStorage)
	handler := profiles.CreateDeleteSkillHandler(invalidAuthRepo{}, storage, skillDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteSkillHandlerBadInput(t *testing.T) {
	storage := new(skillStorage)
	invalidInputFn := func(r *http.Request) (profiles.SkillDeleteInput, error) {
		return profiles.SkillDeleteInput{}, fmt.Errorf("input error")
	}
	handler := profiles.CreateDeleteSkillHandler(&authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteSkillHandlerStorageError(t *testing.T) {
	storage := new(skillStorage)
	storage.On("DeleteSkill", mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("storage error"))
	handler := profiles.CreateDeleteSkillHandler(&authRepo{}, storage, skillDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeleteSkillHandler(t *testing.T) {
	storage := new(skillStorage)
	storage.On("DeleteSkill", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	handler := profiles.CreateDeleteSkillHandler(&authRepo{}, storage, skillDeleteInputFn)
	req, _ := http.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
