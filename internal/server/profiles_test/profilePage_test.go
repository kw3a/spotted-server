package profilestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type profilePageStorage struct {
	mock.Mock
}

func (s *profilePageStorage) GetUser(ctx context.Context, userID string) (database.User, error) {
	args := s.Called(ctx, userID)
	return args.Get(0).(database.User), args.Error(1)
}
func (s *profilePageStorage) SelectEducation(ctx context.Context, userID string) ([]shared.EducationEntry, error) {
	args := s.Called(ctx, userID)
	return args.Get(0).([]shared.EducationEntry), args.Error(1)
}
func (s *profilePageStorage) SelectExperiences(ctx context.Context, userID string) ([]shared.ExperienceEntry, error) {
	args := s.Called(ctx, userID)
	return args.Get(0).([]shared.ExperienceEntry), args.Error(1)
}
func (s *profilePageStorage) SelectLinks(ctx context.Context, userID string) ([]shared.Link, error) {
	args := s.Called(ctx, userID)
	return args.Get(0).([]shared.Link), args.Error(1)
}
func (s *profilePageStorage) SelectParticipatedOffers(ctx context.Context, userID string, page int32) ([]shared.Offer, error) {
	args := s.Called(ctx, userID, page)
	return args.Get(0).([]shared.Offer), args.Error(1)
}
func (s *profilePageStorage) SelectSkills(ctx context.Context, userID string) ([]shared.SkillEntry, error) {
	args := s.Called(ctx, userID)
	return args.Get(0).([]shared.SkillEntry), args.Error(1)
}

func profilePageInputFn(r *http.Request) (profiles.ProfilePageInput, error) {
	return profiles.ProfilePageInput{UserID: "1"}, nil
}

func TestProfilePageHandlerBadAuth(t *testing.T) {
	storage := new(profilePageStorage)
	handler := profiles.CreateProfilePageHandler(invalidAuthRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProfilePageHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.ProfilePageInput, error) {
		return profiles.ProfilePageInput{}, fmt.Errorf("error")
	}
	storage := new(profilePageStorage)
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorage(t *testing.T) {
	inputFn := func(r *http.Request) (profiles.ProfilePageInput, error) {
		return profiles.ProfilePageInput{UserID: "1"}, nil
	}
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, inputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerPage2T(t *testing.T) {
	inputFn := func(r *http.Request) (profiles.ProfilePageInput, error) {
		return profiles.ProfilePageInput{UserID: "1", Page: 2}, nil
	}
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	handler := profiles.CreateProfilePageHandler(authz, &invalidTemplates{}, storage, inputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerPage2(t *testing.T) {
	inputFn := func(r *http.Request) (profiles.ProfilePageInput, error) {
		return profiles.ProfilePageInput{UserID: "1", Page: 2}, nil
	}
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, inputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorageGetUser(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadStorageGetUser(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorageSelectExperiences(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadStorageSelectExperiences(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorageSelectEducation(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadStorageSelectEducation(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorageSelectSkills(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadStorageSelectSkills(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadStorageSelectLinks(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadStorageSelectLinks(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, fmt.Errorf("error"))
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwnerBadTemplate(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, nil)
	handler := profiles.CreateProfilePageHandler(authz, &invalidTemplates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerNotOwnerBadTemplate(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, nil)
	handler := profiles.CreateProfilePageHandler(authRepo{}, &invalidTemplates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProfilePageHandlerOwner(t *testing.T) {
	authz := new(authMock)
	authz.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: auth.AuthRole, ID: "1"}, nil)
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, nil)
	handler := profiles.CreateProfilePageHandler(authz, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
func TestProfilePageHandlerNotOwner(t *testing.T) {
	storage := new(profilePageStorage)
	storage.On("SelectParticipatedOffers", mock.Anything, mock.Anything, mock.Anything).Return([]shared.Offer{}, nil)
	storage.On("GetUser", mock.Anything, mock.Anything).Return(database.User{}, nil)
	storage.On("SelectExperiences", mock.Anything, mock.Anything).Return([]shared.ExperienceEntry{}, nil)
	storage.On("SelectEducation", mock.Anything, mock.Anything).Return([]shared.EducationEntry{}, nil)
	storage.On("SelectSkills", mock.Anything, mock.Anything).Return([]shared.SkillEntry{}, nil)
	storage.On("SelectLinks", mock.Anything, mock.Anything).Return([]shared.Link{}, nil)
	handler := profiles.CreateProfilePageHandler(authRepo{}, &templates{}, storage, profilePageInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
