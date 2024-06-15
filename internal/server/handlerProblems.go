package server

import (
	"context"
	"fmt"
	"net/http"
)

type ProblemsStorage interface {
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
}

func createProblemHandler(templ *Templates, storage ProblemsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		problemID := r.FormValue("problemID")
		if problemID == "" {
			http.Error(w, "invalid problemID", http.StatusBadRequest)
			return
		}
		problem, err := storage.SelectProblem(r.Context(), problemID)
		if err != nil {
			http.Error(w, "problem not found", http.StatusBadRequest)
			return
		}

		err = templ.Render(w, "problem", problem)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't render problem content: %s", err), http.StatusInternalServerError)
		}
	}
}

func (app *App) ProblemsHandler() http.HandlerFunc {
	return createProblemHandler(app.Templ, app.Storage)
}
