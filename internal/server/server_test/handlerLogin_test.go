package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
	"github.com/stretchr/testify/mock"
)

func loginInputFn(r *http.Request) (server.LoginInput, error) {
	return server.LoginInput{}, nil
}

type loginStorage struct {
	mock.Mock
}

func (s *loginStorage) GetUserID(ctx context.Context, email, password string) (string, error) {
	args := s.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (s *loginStorage) Save(ctx context.Context, refreshToken string) error {
	args := s.Called(ctx, refreshToken)
	return args.Error(0)
}

type loginAuthType struct {
	mock.Mock
}

func (t *loginAuthType) CreateTokens(userID string) (refresh string, access string, err error) {
	args := t.Called(userID)
	return args.String(0), args.String(1), args.Error(2)
}

func TestLoginInputEmptyEmail(t *testing.T) {
	formValues := map[string][]string{
		"email":    {""},
		"password": {"mypassword"},
	}
	req := formRequest("POST", "/", formValues)
	_, err := server.GetLoginInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoginInputBadEmail(t *testing.T) {
	formValues := map[string][]string{
		"email":    {"email"},
		"password": {"mypassword"},
	}
	req := formRequest("POST", "/", formValues)
	_, err := server.GetLoginInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoginInputEmptyPassword(t *testing.T) {
	formValues := map[string][]string{
		"email":    {"myemail@mail.com"},
		"password": {""},
	}
	req := formRequest("POST", "/", formValues)
	_, err := server.GetLoginInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoginInputBadPassword(t *testing.T) {
	formValues := map[string][]string{
		"email":    {"myemail@mail.com"},
		"password": {"exceeds30characterspasswordexceeds30characterspassword"},
	}
	req := formRequest("POST", "/", formValues)
	_, err := server.GetLoginInput(req)
	if err == nil {
		t.Error(err)
	}
}

func TestLoginInput(t *testing.T) {
	formValues := map[string][]string{
		"email":    {"myemail@mail.com"},
		"password": {"mypassword"},
	}
	req := formRequest("POST", "/", formValues)
	input, err := server.GetLoginInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.Email != "myemail@mail.com" {
		t.Error("does not match email")
	}
	if input.Password != "mypassword" {
		t.Error("does not match password")
	}
}


func TestLoginHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.LoginInput, error) {
		return server.LoginInput{}, errors.New("error")
	}
	storage := new(loginStorage)
	authType := new(loginAuthType)
	handler := server.CreateLoginHandler(authType, storage, invalidInputFn)
	req, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginHandlerBadStorageGetUserID(t *testing.T) {
	storage := new(loginStorage)
	storage.On("GetUserID", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("error"))
	authType := new(loginAuthType)
	handler := server.CreateLoginHandler(authType, storage, loginInputFn)
	req, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest{
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginHandlerBadAuthTypeCreateTokens(t *testing.T) {
	storage := new(loginStorage)
	storage.On("GetUserID", mock.Anything, mock.Anything, mock.Anything).Return("1", nil)
	authType := new(loginAuthType)
	authType.On("CreateTokens", mock.Anything).Return("", "", errors.New("error"))
	handler := server.CreateLoginHandler(authType, storage, loginInputFn)
	req, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLoginHandlerBadStorageSave(t *testing.T) {
	storage := new(loginStorage)
	storage.On("GetUserID", mock.Anything, mock.Anything, mock.Anything).Return("1", nil)
	authType := new(loginAuthType)
	authType.On("CreateTokens", mock.Anything).Return("1", "2", nil)
	storage.On("Save", mock.Anything, mock.Anything).Return(errors.New("error"))
	handler := server.CreateLoginHandler(authType, storage, loginInputFn)
	req, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLoginHandler(t *testing.T) {
	storage := new(loginStorage)
	storage.On("GetUserID", mock.Anything, mock.Anything, mock.Anything).Return("1", nil)
	authType := new(loginAuthType)
	authType.On("CreateTokens", mock.Anything).Return("1", "2", nil)
	storage.On("Save", mock.Anything, mock.Anything).Return(nil)
	handler := server.CreateLoginHandler(authType, storage, loginInputFn)
	req, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("HX-Redirect") != "/" {
		t.Errorf("expected redirect %s, got %s", "/", w.Header().Get("HX-Redirect"))
	}
}
