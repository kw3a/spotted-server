package server

import (
	"context"
	"net/http"
)

type ScoreStorage interface {
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
}

func createScoreHandler(templ *Templates, storage ScoreStorage, authService AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		problemID := r.FormValue("problemID")
		if problemID == "" {
			http.Error(w, "problemID is empty", http.StatusBadRequest)
			return
		}
		score, err := storage.SelectScore(r.Context(), userID, problemID)
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
	return createScoreHandler(
		app.Templ,
		app.Storage, 
    app.AuthService,
    )
}
