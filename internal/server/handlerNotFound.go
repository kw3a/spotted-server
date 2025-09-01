package server

import "net/http"

func CreateNotFoundHandler(tmpl TemplatesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err :=tmpl.Render(w, "notFound", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (DI *App) NotFoundHandler() http.HandlerFunc {
	return CreateNotFoundHandler(DI.Templ)
}
