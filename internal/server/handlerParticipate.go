package server

import (
	"context"
	"net/http"
)

type ParticipationStorage interface {
	Participate(ctx context.Context, userID string, quizID string) error
}

func createParticipateHandler(storage ParticipationStorage, authService AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quizID := r.FormValue("quizID")

		if userID == "" || quizID == "" {
			http.Error(w, "invalid user id or quiz id", http.StatusBadRequest)
			return
		}
		err = storage.Participate(r.Context(), userID, quizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("HX-Redirect", "/quizzes/"+quizID)
		w.WriteHeader(http.StatusOK)
	}
}

func (DI *App) ParticipateHandler() http.HandlerFunc {
	return createParticipateHandler(DI.Storage, DI.AuthService)
}
