package server

import (
	"context"
	"fmt"
	"net/http"
)

type ExamplesStorage interface {
	SelectExamples(ctx context.Context, problemID string) ([]Example, error)
}

func createExamplesHandler(templ *Templates, storage ExamplesStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		problemID := r.FormValue("problemID")
		if problemID == "" {
			http.Error(w, "invalid exampleID", http.StatusBadRequest)
			return
		}
		examples, err := storage.SelectExamples(r.Context(), problemID)
		if err != nil {
			http.Error(w, "example not found", http.StatusBadRequest)
			return
		}

		err = templ.Render(w, "examples", examples)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't render example content: %s", err), http.StatusInternalServerError)
		}
	}
}

func (app *App) ExamplesHandler() http.HandlerFunc {
	return createExamplesHandler(app.Templ, app.Storage)
}
