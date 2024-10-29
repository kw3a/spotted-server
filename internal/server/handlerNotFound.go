package server

import "net/http"

func CreateNotFoundHandler(tmpl TemplatesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Render(w, "notFound", nil)
	}
}

func (DI *App) NotFoundHandler() http.HandlerFunc {
	return CreateNotFoundHandler(DI.Templ)
}
