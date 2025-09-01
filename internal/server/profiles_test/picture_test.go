package profilestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
)

type pictureStorage struct {
	mock.Mock
}

func (s *pictureStorage) UpdateProfilePic(ctx context.Context, userID string, imageURL string) error {
	args := s.Called(ctx, userID, imageURL)
	return args.Error(0)
}

func pictureInputFn(
	r *http.Request,
	cloudinaryService shared.CloudinaryService,
) (profiles.ProfilePicInput, error) {
	return profiles.ProfilePicInput{}, nil
}

func TestPictureHandlerBadAuth(t *testing.T) {
	storage := new(pictureStorage)
	cld := new(cldMock)
	handler := profiles.CreatePictureHandler(storage, &invalidAuthRepo{}, cld, pictureInputFn)
	req, _ := http.NewRequest("PATCH", "/pictures", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
	}
}

func TestPictureHandlerVisitor(t *testing.T) {
	a := new(authMock)
	a.On("GetUser", mock.Anything).Return(auth.AuthUser{Role: "visitor"}, nil)
	storage := new(pictureStorage)
	storage.On("UpdateProfilePic", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cld := new(cldMock)
	handler := profiles.CreatePictureHandler(storage, a, cld, pictureInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
	}
}

func TestPictureHandlerBadInput(t *testing.T) {
	pictureInputFn := func(
		r *http.Request,
		cloudinaryService shared.CloudinaryService,
	) (profiles.ProfilePicInput, error) {
		return profiles.ProfilePicInput{}, fmt.Errorf("input error")
	}
	storage := new(pictureStorage)
	cld := new(cldMock)
	handler := profiles.CreatePictureHandler(storage, &authRepo{}, cld, pictureInputFn)
	req, _ := http.NewRequest("PATCH", "/pictures", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected %v, got %v", http.StatusBadRequest, w.Code)
	}
}

func TestPictureHandlerBadStorage(t *testing.T) {
	storage := new(pictureStorage)
	storage.On("UpdateProfilePic", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	cld := new(cldMock)
	handler := profiles.CreatePictureHandler(storage, &authRepo{}, cld, pictureInputFn)
	req, _ := http.NewRequest("PATCH", "/pictures", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, w.Code)
	}
}

func TestPictureHandler(t *testing.T) {
	storage := new(pictureStorage)
	storage.On("UpdateProfilePic", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cld := new(cldMock)
	handler := profiles.CreatePictureHandler(storage, &authRepo{}, cld, pictureInputFn)
	req, _ := http.NewRequest("PATCH", "/pictures", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Trigger") != "image-changed" {
		t.Errorf("expected redirect %s, got %s", "image.changed", w.Header().Get("HX-Trigger"))
	}
}
