package urlParams

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ProblemID(r *http.Request) (string, error) {
	problemID := chi.URLParam(r, "problemID")
	err := uuid.Validate(problemID)
	if err != nil {
		return "", err
	}
	return problemID, nil
}
