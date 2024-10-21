package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
)

type LoginPageStorage interface {
}

type LoginPageData struct {
	User auth.AuthUser
}

func CreateLoginPageHandler(authService AuthRep, templ TemplatesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		err = templ.Render(w, "loginPage", LoginPageData{
			User: user,
		})
		if err != nil {
			http.Error(w, "can't render login page", http.StatusInternalServerError)
		}
	}
}

func (DI *App) LoginPageHandler() http.HandlerFunc {
	return CreateLoginPageHandler(DI.AuthService, DI.Templ)
}
