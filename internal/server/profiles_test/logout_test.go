package profilestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type mockLogoutStorage struct {
	mock.Mock
}

func (m *mockLogoutStorage) Revoke(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func TestLogoutHandlerWithCookie(t *testing.T) {
	path := "/login"
	storage := new(mockLogoutStorage)
	storage.On("Revoke", mock.Anything, "test-refresh-token").Return(nil)
	handler := profiles.CreateLogoutHandler(storage, path)
	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "test-refresh-token"})
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
	hxRedirect := w.Header().Get("HX-Redirect")
	if hxRedirect != path {
		t.Errorf("expected HX-Redirect header to be '/login', got '%s'", hxRedirect)
	}
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
			if cookie.MaxAge != -1 {
				t.Errorf("expected cookie %s to be deleted (MaxAge = -1), got MaxAge = %d", cookie.Name, cookie.MaxAge)
			}
		}
	}
	storage.AssertExpectations(t)
}

func TestLogoutHandlerWithoutCookie(t *testing.T) {
	path:="/login"
	storage := new(mockLogoutStorage)
	handler := profiles.CreateLogoutHandler(storage, path)
	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
	hxRedirect := w.Header().Get("HX-Redirect")
	if hxRedirect != path {
		t.Errorf("expected HX-Redirect header to be %s, got %s", path, hxRedirect)
	}
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
			if cookie.MaxAge != -1 {
				t.Errorf("expected cookie %s to be deleted (MaxAge = -1), got MaxAge = %d", cookie.Name, cookie.MaxAge)
			}
		}
	}
	storage.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything)
}

func TestLogoutHandlerStorageError(t *testing.T) {
	path := "/login"
	storage := new(mockLogoutStorage)
	storage.On("Revoke", mock.Anything, "test-refresh-token").Return(errors.New("storage error"))
	handler := profiles.CreateLogoutHandler(storage, path)
	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "test-refresh-token"})
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
	hxRedirect := w.Header().Get("HX-Redirect")
	if hxRedirect != path {
		t.Errorf("expected HX-Redirect header to be %s, got %s", path, hxRedirect)
	}
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
			if cookie.MaxAge != -1 {
				t.Errorf("expected cookie %s to be deleted (MaxAge = -1), got MaxAge = %d", cookie.Name, cookie.MaxAge)
			}
		}
	}
	storage.AssertExpectations(t)
}
