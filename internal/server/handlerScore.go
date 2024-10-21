package server

import (
	"context"
	"net/http"
)

type ScoreStorage interface {
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
}
type ScoreInput struct {
	ProblemID string
}
func GetScoreInput(r *http.Request) (ScoreInput, error) {
	problemID := r.FormValue("problemID")
	if err := ValidateUUID(problemID); err != nil {
		return ScoreInput{}, err
	}
	return ScoreInput{
		ProblemID: problemID,
	}, nil
}
type scoreInputFn func(r *http.Request) (ScoreInput, error)
func CreateScoreHandler(templ TemplatesRepo, storage ScoreStorage, authService AuthRep, inputFn scoreInputFn) http.HandlerFunc {
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

func (app *App) ScoreHandler() http.HandlerFunc {
	return CreateScoreHandler(
		app.Templ,
		app.Storage, 
    app.AuthService,
		GetScoreInput,
    )
}
