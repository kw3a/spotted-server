package quizes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ProblemsStorage interface {
	SelectProblem(ctx context.Context, problemID string) (shared.Problem, error)
}
type ProblemsInput struct {
	ProblemID string
}

func GetProblemsInput(r *http.Request) (ProblemsInput, error) {
	problemID := r.FormValue("problemID")
	if err := shared.ValidateUUID(problemID); err != nil {
		return ProblemsInput{}, err
	}
	return ProblemsInput{
		ProblemID: problemID,
	}, nil
}

type problemsInputFn func(r *http.Request) (ProblemsInput, error)
func CreateProblemHandler(templ shared.TemplatesRepo, storage ProblemsStorage, inputFn problemsInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		problem, err := storage.SelectProblem(r.Context(), input.ProblemID)
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

