package profiles

import (
	"context"
	"net/http"
	"regexp"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errEmailInvalid = "Correo inválido"
	errPasswordLength = "Debe tener entre 5 a 30 caracteres"
	errNotFound = "El correo y la contraseña no coinciden"
	errUnexpected = "Error inesperado, inténtelo de nuevo"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginErr struct {
	EmailErr    string
	PasswordErr string
}

func EmailValidation(email string) string {
	exp := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	compiledRegExp := regexp.MustCompile(exp)
	if !compiledRegExp.MatchString(email) {
		return errEmailInvalid
	}
	return ""
}

func GetLoginInput(r *http.Request) (LoginInput, LoginErr, bool) {
	errFound := false
	loginErr := LoginErr{}
	email := r.FormValue("email")
	password := r.FormValue("password")
	if strErr := EmailValidation(email); strErr != "" {
		loginErr.EmailErr = strErr
		errFound = true
	}
	if len(password) < 5 || len(password) > 30 {
		loginErr.PasswordErr = shared.ErrLength(5, 30)
		errFound = true
	}
	return LoginInput{
		Email:    email,
		Password: password,
	}, loginErr, errFound
}

type LoginStorage interface {
	GetUserID(ctx context.Context, email, password string) (string, error)
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
		userID, err := storage.GetUserID(r.Context(), input.Email, input.Password)
		if err != nil {
			inputErrors.EmailErr = errNotFound
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		refreshToken, accessToken, err := authType.CreateTokens(userID)
		if err != nil {
			inputErrors.EmailErr = errUnexpected
			if err := templ.Render(w, "loginFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.Save(r.Context(), refreshToken)
		if err != nil {
			inputErrors.EmailErr = errUnexpected
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
