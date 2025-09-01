package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
	if user.ID != "123" {
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
