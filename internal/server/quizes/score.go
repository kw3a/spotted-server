package quizes

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ScoreStorage interface {
	SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error)
}
type ScoreInput struct {
	ProblemID string
}

func GetScoreInput(r *http.Request) (ScoreInput, error) {
	problemID := r.FormValue("problemID")
	if err := shared.ValidateUUID(problemID); err != nil {
		return ScoreInput{}, err
	}
	return ScoreInput{
		ProblemID: problemID,
	}, nil
}

type scoreInputFn func(r *http.Request) (ScoreInput, error)
func CreateScoreHandler(templ shared.TemplatesRepo, storage ScoreStorage, authService shared.AuthRep, inputFn scoreInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		score, err := storage.SelectScore(r.Context(), user.ID, input.ProblemID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = templ.Render(w, "score", score)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
