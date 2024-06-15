package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type LanguageStorage interface {
	SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error)
}

func createLanguagesHandler(templ *Templates, storage LanguageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizID := chi.URLParam(r, "quizID")
		if quizID == "" {
			http.Error(w, "invalid quiz id", http.StatusBadRequest)
			return
		}
		languages, err := storage.SelectLanguages(r.Context(), quizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = templ.Render(w, "languages", languages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (DI *App) LanguagesHandler() http.HandlerFunc {
	return createLanguagesHandler(DI.Templ, DI.Storage)
}
