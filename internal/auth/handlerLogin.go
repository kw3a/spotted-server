package auth

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

type LoginInput struct {
	Email    string
	Password string
}

func (input LoginInput) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if input.Email == "" {
		problems["email_error"] = "email cannot be empty"
	}
	exp := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	compiledRegExp := regexp.MustCompile(exp)
	if !compiledRegExp.MatchString(input.Email) {
		problems["email_error"] = "invalid email address"
	}
	if len(input.Password) == 0 || len(input.Password) > 30 {
		problems["password_error"] = "password length must be less than 30 and non empty"
	}
	return problems
}

func DecodeLoginInput(r *http.Request) (LoginInput, map[string]string) {
	input := LoginInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	return input, input.Valid(r.Context())
}

func CreateLoginHandler(authRep Authentication, storage AuthenticationStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, problems := DecodeLoginInput(r)
		if len(problems) > 0 {
			http.Error(w, fmt.Sprintf("%v", problems), http.StatusBadRequest)
			return
		}
		userID, err := storage.GetUserID(r.Context(), AuthStorageParams(input))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		refreshToken, err := authRep.CreateRefresh(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		accessToken, err := authRep.CreateAccess(refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = storage.Save(r.Context(), refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.Header().Set("HX-Location", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func (auth *AuthService) LoginHandler() http.HandlerFunc {
	return CreateLoginHandler(
		auth.JWTRep,
		auth.Storage,
	)
}
