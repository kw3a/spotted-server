package server

import (
	"context"
	"fmt"
	"net/http"
)

type ExamplesStorage interface {
	SelectExamples(ctx context.Context, problemID string) ([]Example, error)
}

type ExamplesInput struct {
	ProblemID string
}

func GetExamplesInput(r *http.Request) (ExamplesInput, error) {
	problemID := r.FormValue("problemID")
	if err:= ValidateUUID(problemID); err != nil {
		return ExamplesInput{}, fmt.Errorf("problemID is not a valid UUID")
	}
	return ExamplesInput{
		ProblemID: problemID,
	}, nil
}
type examplesInputFunc = func(r *http.Request) (ExamplesInput, error)
func CreateExamplesHandler(templ TemplatesRepo, storage ExamplesStorage, inputFn examplesInputFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		examples, err := storage.SelectExamples(r.Context(), input.ProblemID)
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
	return CreateExamplesHandler(app.Templ, app.Storage, GetExamplesInput)
}
