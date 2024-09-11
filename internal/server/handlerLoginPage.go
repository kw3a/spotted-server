package server

import "net/http"

type LoginPageStorage interface {
}

func CreateLoginPageHandler(templ TemplatesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templ.Render(w, "loginPage", "")
		if err != nil {
			http.Error(w, "can't render login page", http.StatusInternalServerError)
		}
	}
}

func (DI *App) LoginPageHandler() http.HandlerFunc {
	return CreateLoginPageHandler(DI.Templ)
}
