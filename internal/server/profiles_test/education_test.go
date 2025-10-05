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

type educationStorage struct {
	mock.Mock
}

func (s *educationStorage) CountEducation(ctx context.Context, userID string) (int32, error) {
	args := s.Called(ctx, userID)
	return int32(args.Int(0)), args.Error(1)
}

func (s *educationStorage) DeleteEducation(ctx context.Context, userID string, educationID string) error {
	args := s.Called(ctx, userID, educationID)
	return args.Error(0)
}
func (s *educationStorage) RegisterEducation(
	ctx context.Context,
	educationID string,
	userID string,
	institution string,
	degree string,
	start time.Time,
	end sql.NullTime,
) error {
	args := s.Called(ctx, educationID, userID, institution, degree, start, end)
	return args.Error(0)
}

func edInputFn(r *http.Request) (profiles.EducationRegisterInput, profiles.EducationRegErrors, bool) {
	return profiles.EducationRegisterInput{}, profiles.EducationRegErrors{}, false
}

func edDeleteInputFn(r *http.Request) (profiles.EducationDeleteInput, error) {
	return profiles.EducationDeleteInput{}, nil
}

func TestEdHandlerBadAuth(t *testing.T) {
	storage := new(educationStorage)
	handler := profiles.CreateRegisterEducationHandler(&templates{}, invalidAuthRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestEdHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.EducationRegisterInput, profiles.EducationRegErrors, bool) {
		return profiles.EducationRegisterInput{}, profiles.EducationRegErrors{}, true
	}
	storage := new(educationStorage)
	handler := profiles.CreateRegisterEducationHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEdHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.EducationRegisterInput, profiles.EducationRegErrors, bool) {
		return profiles.EducationRegisterInput{}, profiles.EducationRegErrors{}, true
	}
	storage := new(educationStorage)
	handler := profiles.CreateRegisterEducationHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEdHandlerBadStorageCountEducationT(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, fmt.Errorf("count error"))
	handler := profiles.CreateRegisterEducationHandler(&invalidTemplates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEdHandlerBadStorageCountEducation(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, fmt.Errorf("count error"))
	handler := profiles.CreateRegisterEducationHandler(&templates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEdHandlerBadStorageRegisterT(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, nil)
	storage.On(
		"RegisterEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("ed register error"))
	handler := profiles.CreateRegisterEducationHandler(&invalidTemplates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEdHandlerBadStorageRegister(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, nil)
	storage.On(
		"RegisterEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("ed register error"))
	handler := profiles.CreateRegisterEducationHandler(&templates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEdHandlerBadTemplate(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, nil)
	storage.On(
		"RegisterEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateRegisterEducationHandler(&invalidTemplates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "ed-added" {
		t.Errorf("expected redirect %s, got %s", "ed-added", w.Header().Get("HX-Trigger"))
	}
}

func TestEdHandler(t *testing.T) {
	storage := new(educationStorage)
	storage.On("CountEducation", mock.Anything, mock.Anything).Return(0, nil)
	storage.On(
		"RegisterEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateRegisterEducationHandler(&templates{}, &authRepo{}, storage, edInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "ed-added" {
		t.Errorf("expected redirect %s, got %s", "ed-added", w.Header().Get("HX-Trigger"))
	}
}

func TestEdDeleteHandlerBadAuth(t *testing.T) {
	storage := new(educationStorage)
	handler := profiles.CreateDeleteEducationHandler(&invalidAuthRepo{}, storage, edDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestEdDeleteHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.EducationDeleteInput, error) {
		return profiles.EducationDeleteInput{}, fmt.Errorf("error")
	}
	storage := new(educationStorage)
	handler := profiles.CreateDeleteEducationHandler(&authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEdDeleteHandlerBadStorage(t *testing.T) {
	storage := new(educationStorage)
	storage.On(
		"DeleteEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("error"))
	handler := profiles.CreateDeleteEducationHandler(&authRepo{}, storage, edDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestEdDeleteHandler(t *testing.T) {
	storage := new(educationStorage)
	storage.On(
		"DeleteEducation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	handler := profiles.CreateDeleteEducationHandler(&authRepo{}, storage, edDeleteInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
