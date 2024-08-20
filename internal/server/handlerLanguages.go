package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type LanguageStorage interface {
	SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error)
}

type LanguagesInput struct {
	QuizID string
}

func GetLanguagesInput(r *http.Request) (LanguagesInput, error) {
	quizID := chi.URLParam(r, "quizID")
	if err := ValidateUUID(quizID); err != nil {
		return LanguagesInput{}, err
	}
	return LanguagesInput{
		QuizID: quizID,
	}, nil
}

type languagesInputFn func(r *http.Request) (LanguagesInput, error)
func CreateLanguagesHandler(templ TemplatesRepo, storage LanguageStorage, inputFn languagesInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		languages, err := storage.SelectLanguages(r.Context(), input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = templ.Render(w, "languages", languages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (DI *App) LanguagesHandler() http.HandlerFunc {
	return CreateLanguagesHandler(DI.Templ, DI.Storage, GetLanguagesInput)
}
