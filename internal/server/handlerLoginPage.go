package server

import "net/http"

type LoginPageStorage interface {
}

func createLoginPageHandler(templ *Templates) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templ.Render(w, "loginPage", "")
		if err != nil {
			http.Error(w, "can't render login page", http.StatusInternalServerError)
		}
	}
}

func (DI *App) LoginPageHandler() http.HandlerFunc {
		return createLoginPageHandler(DI.Templ)
}
