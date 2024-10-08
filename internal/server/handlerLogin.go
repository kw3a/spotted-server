package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/kw3a/spotted-server/internal/auth"
)

type LoginInput struct {
	Email    string
	Password string
}

func GetLoginInput(r *http.Request) (LoginInput, error) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" {
		return LoginInput{}, fmt.Errorf("email cannot be empty")
	}
	exp := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	compiledRegExp := regexp.MustCompile(exp)
	if !compiledRegExp.MatchString(email) {
		return LoginInput{}, fmt.Errorf("invalid email address")
	}
	if len(password) == 0 || len(password) > 30 {
		return LoginInput{}, fmt.Errorf("password length must be less than 30 and non empty")
	}
	return LoginInput{
		Email:    email,
		Password: password,
	}, nil
}
type LoginStorage interface {
	GetUserID(ctx context.Context, email, password string) (string, error)
	Save(ctx context.Context, refreshToken string) error
}
type LoginAuthType interface {
	CreateTokens(userID string) (refresh string, access string, err error)
}
type loginInputFn func(r *http.Request) (LoginInput, error)

func CreateLoginHandler(authType LoginAuthType, storage LoginStorage, inputFn loginInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userID, err := storage.GetUserID(r.Context(), input.Email, input.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		refreshToken, accessToken, err := authType.CreateTokens(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = storage.Save(r.Context(), refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		auth.SetTokenCookie(w, "refresh_token", refreshToken)
		auth.SetTokenCookie(w, "access_token", accessToken)
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}
func (DI *App) LoginHandler() http.HandlerFunc {
	return CreateLoginHandler(
		DI.AuthType,
		DI.Storage,
		GetLoginInput,
	)
}
