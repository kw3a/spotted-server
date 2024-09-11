package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type middlewareStorage struct{}
func (m *middlewareStorage) GetRole(ctx context.Context, userID string) (string, error) {
	return "admin", nil
}

type invalidMiddlewareStorage struct{}
func (i *invalidMiddlewareStorage) GetRole(ctx context.Context, userID string) (string, error) {
	return "", errors.New("error")
}

type middlewareAuthType struct{}
func (m *middlewareAuthType) CreateAccess(refreshToken string) (string, error) {
	return "", nil
}
func (m *middlewareAuthType) WhoAmI(accessToken string) (userID string, err error) {
	return "", nil
}

type invalidMiddlewareAuthType struct{}
func (i *invalidMiddlewareAuthType) CreateAccess(refreshToken string) (string, error) {
	return "", errors.New("error")
}
func (i *invalidMiddlewareAuthType) WhoAmI(accessToken string) (userID string, err error) {
	return "", errors.New("error")
}

func TestGetUserEmpty(t *testing.T) {
	authService := AuthService{}
	r, _ := http.NewRequest("GET", "/", nil)
	_, err := authService.GetUser(r)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestGetUser(t *testing.T) {
	authService := AuthService{}
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, AuthUser{}, AuthUser{ID: "123", Role: "admin"})
	r = r.WithContext(ctx)
	user, err := authService.GetUser(r)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if user != "123" {
		t.Errorf("expected 123, got %v", user)
	}
}

func TestGetCookiesEmpty(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	_, _, err := getCookies(r)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestGetCookies(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "access"})
	r.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
	refresh, access, err := getCookies(r)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if refresh != "refresh" {
		t.Errorf("expected refresh, got %v", refresh)
	}
	if access != "access" {
		t.Errorf("expected access, got %v", access)
	}
}

func TestSetTokenCookie(t *testing.T) {
	w := httptest.NewRecorder()
	cookie_name := "cookie_name"
	cookie_value := "cookie_value"
	SetTokenCookie(w, cookie_name, cookie_value)
	cookies := w.Result().Cookies()
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cookie_name {
			cookie = c
		}
	}
	if cookie == nil {
		t.Errorf("expected cookie, got nil")
	} else {
		if cookie.Value != cookie_value {
			t.Errorf("expected %v, got %v", cookie_value, cookie.Value)
		}
	}
}

func TestDeleteTokenCookie(t *testing.T) {
	w := httptest.NewRecorder()
	cookie_name := "cookie_name"
	cookie_value := "cookie_value"
	http.SetCookie(w, &http.Cookie{Name: cookie_name, Value: cookie_value})
	deleteTokenCookie(w, cookie_name)
	cookies := w.Result().Cookies()
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cookie_name {
			cookie = c
		}
	}
	if cookie == nil {
		t.Errorf("expected cookie, got nil")
	} else {
		if cookie.Value != "" {
			t.Errorf("expected empty, got %v", cookie.Value)
		}
		if cookie.MaxAge != -1 {
			t.Errorf("expected -1, got %v", cookie.MaxAge)
		}
		if !cookie.Expires.Before(time.Now()) {
			t.Errorf("expected expired cookie, got %v", cookie.Expires)
		}
	}
}

func TestDeleteCookies(t *testing.T) {
	w := httptest.NewRecorder()
	DeleteCookies(w)
	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Errorf("expected 2 cookies, got %v", len(cookies))
	}
	for _, cookie := range cookies {
		if cookie.Value != "" {
			t.Errorf("expected empty, got %v", cookie.Value)
		}
		if cookie.MaxAge != -1 {
			t.Errorf("expected -1, got %v", cookie.MaxAge)
		}
		if !cookie.Expires.Before(time.Now()) {
			t.Errorf("expected expired cookie, got %v", cookie.Expires)
		}
	}
}

func nextFn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
func requestWithCookies() *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "access"})
	r.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
	return r
}

func TestCreateMiddlewareEmptyCookies(t *testing.T) {
	middleware := CreateMiddleware(&middlewareStorage{}, &middlewareAuthType{}, "/", "admin", nextFn())
	if middleware == nil {
		t.Errorf("expected middleware, got nil")
	}
	req, _:= http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected %v, got %v", http.StatusSeeOther, w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Errorf("expected /, got %v", w.Header().Get("Location"))
	}
}

func TestCreateMiddlewareBadStorage(t *testing.T) {
	middleware := CreateMiddleware(&invalidMiddlewareStorage{}, &middlewareAuthType{}, "/", "admin", nextFn())
	if middleware == nil {
		t.Errorf("expected middleware, got nil")
	}
	req := requestWithCookies()
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected %v, got %v", http.StatusSeeOther, w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Errorf("expected /, got %v", w.Header().Get("Location"))
	}
}

func TestCreateMiddlewareBadRole(t *testing.T) {
	middleware := CreateMiddleware(&middlewareStorage{}, &middlewareAuthType{}, "/", "user", nextFn())
	if middleware == nil {
		t.Errorf("expected middleware, got nil")
	}
	req := requestWithCookies()
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected %v, got %v", http.StatusSeeOther, w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Errorf("expected /, got %v", w.Header().Get("Location"))
	}
}

func TestCreateMiddlewareBadAuthType(t *testing.T) {
	middleware := CreateMiddleware(&middlewareStorage{}, &invalidMiddlewareAuthType{}, "/", "user", nextFn())
	if middleware == nil {
		t.Errorf("expected middleware, got nil")
	}
	req := requestWithCookies()
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected %v, got %v", http.StatusSeeOther, w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Errorf("expected /, got %v", w.Header().Get("Location"))
	}
}

func TestCreateMiddleware(t *testing.T) {
	middleware := CreateMiddleware(&middlewareStorage{}, &middlewareAuthType{}, "/", "admin", nextFn())
	if middleware == nil {
		t.Errorf("expected middleware, got nil")
	}
	req := requestWithCookies()
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, w.Code)
	}
	res := w.Body.String()
	if res != "OK" {
		t.Errorf("expected OK, got %v", res)
	}
}
