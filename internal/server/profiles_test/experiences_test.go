package profilestest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type expStorage struct {
	mock.Mock
}

func (s *expStorage) DeleteExperience(ctx context.Context, userID string, experienceID string) error {
	args := s.Called(ctx, userID, experienceID)
	return args.Error(0)
}
func (s *expStorage) RegisterExperience(
	ctx context.Context,
	experienceID string,
	userID string,
	company string,
	title string,
	start time.Time,
	end sql.NullTime,
) error {
	args := s.Called(ctx, experienceID, userID, company, title, start, end)
	return args.Error(0)
}

func expInputFn(r *http.Request) (profiles.ExpRegInput, profiles.ExpRegErrors, bool) {
	return profiles.ExpRegInput{}, profiles.ExpRegErrors{}, false
}

func expDeleteInputFn(r *http.Request) (profiles.ExperienceDeleteInput, error) {
	return profiles.ExperienceDeleteInput{}, nil
}

func TestExpHandlerBadAuth(t *testing.T) {
	storage := new(expStorage)
	handler := profiles.CreateRegisterExperienceHandler(&templates{}, invalidAuthRepo{}, storage, expInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.ExpRegInput, profiles.ExpRegErrors, bool) {
		return profiles.ExpRegInput{}, profiles.ExpRegErrors{}, true
	}
	storage := new(expStorage)
	handler := profiles.CreateRegisterExperienceHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.ExpRegInput, profiles.ExpRegErrors, bool) {
		return profiles.ExpRegInput{}, profiles.ExpRegErrors{}, true
	}
	storage := new(expStorage)
	handler := profiles.CreateRegisterExperienceHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpHandlerBadStorage(t *testing.T) {
	storage := new(expStorage)
	storage.On(
		"RegisterExperience",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := profiles.CreateRegisterExperienceHandler(&templates{}, &authRepo{}, storage, expInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpHandlerBadTemplate(t *testing.T) {
	storage := new(expStorage)
	storage.On(
		"RegisterExperience",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateRegisterExperienceHandler(&invalidTemplates{}, &authRepo{}, storage, expInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "exp-added" {
		t.Errorf("expected redirect %s, got %s", "exp-added", w.Header().Get("HX-Trigger"))
	}
}

func TestExpHandler(t *testing.T) {
	storage := new(expStorage)
	storage.On(
		"RegisterExperience",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateRegisterExperienceHandler(&templates{}, &authRepo{}, storage, expInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "exp-added" {
		t.Errorf("expected redirect %s, got %s", "exp-added", w.Header().Get("HX-Trigger"))
	}
}

func TestExpDeleteHandlerBadAuth(t *testing.T) {
	storage := new(expStorage)
	handler := profiles.CreateDeleteExperienceHandler(&invalidAuthRepo{}, storage, expDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpDeleteHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.ExperienceDeleteInput, error) {
		return profiles.ExperienceDeleteInput{}, fmt.Errorf("error")
	}
	storage := new(expStorage)
	handler := profiles.CreateDeleteExperienceHandler(&authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpDeleteHandlerBadStorage(t *testing.T) {
	storage := new(expStorage)
	storage.On(
		"DeleteExperience",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := profiles.CreateDeleteExperienceHandler(&authRepo{}, storage, expDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpDeleteHandler(t *testing.T) {
	storage := new(expStorage)
	storage.On(
		"DeleteExperience",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateDeleteExperienceHandler(&authRepo{}, storage, expDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK{
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
