package profiles

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errEmailInvalid   = "Correo inválido"
	errUnexpected     = "Error inesperado, inténtelo de nuevo"
)

type LoginInput struct {
	Nick     string
	Password string
}

type LoginErr struct {
	NickErr     string
	PasswordErr string
}

func GetLoginInput(r *http.Request) (LoginInput, LoginErr, bool) {
	errFound := false
	loginErr := LoginErr{}
	nick := r.FormValue("nick")
	password := r.FormValue("password")
	if len(nick) < 3 || len(nick) > 32 {
		loginErr.NickErr = shared.ErrLength(3, 32)
		errFound = true
	}
	if len(password) < 5 || len(password) > 30 {
		loginErr.PasswordErr = shared.ErrLength(5, 30)
		errFound = true
	}
	return LoginInput{
		Nick:     nick,
		Password: password,
	}, loginErr, errFound
}

type LoginStorage interface {
	GetUserID(ctx context.Context, nick, password string) (string, error)
	Save(ctx context.Context, refreshToken string) error
}
type LoginAuthType interface {
	CreateTokens(userID string) (refresh string, access string, err error)
}
type loginInputFn func(r *http.Request) (LoginInput, LoginErr, bool)

func CreateLoginHandler(
	authType LoginAuthType,
	storage LoginStorage,
	inputFn loginInputFn,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, inputErrors, errorExists := inputFn(r)
		if errorExists {
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		userID, err := storage.GetUserID(r.Context(), input.Nick, input.Password)
		if err != nil {
			inputErrors.NickErr = err.Error()
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		refreshToken, accessToken, err := authType.CreateTokens(userID)
		if err != nil {
			inputErrors.NickErr = errUnexpected
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.Save(r.Context(), refreshToken)
		if err != nil {
			inputErrors.NickErr = errUnexpected
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		auth.SetTokenCookie(w, "refresh_token", refreshToken)
		auth.SetTokenCookie(w, "access_token", accessToken)
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}
