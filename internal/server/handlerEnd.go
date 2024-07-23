package server

import (
	"context"
	"net/http"
)

type EndStorage interface {
	EndQuiz(ctx context.Context, userID, quizID string) error
}

	
func createEndHandler(endStorage EndStorage, authService AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quizID := r.FormValue("quizID")

		if quizID == "" {
			http.Error(w, "invalid quiz id", http.StatusBadRequest)
			return
		}
		err = endStorage.EndQuiz(r.Context(), user, quizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func (DI *App) EndHandler() http.HandlerFunc {
	return createEndHandler(DI.Storage, DI.AuthService)
}
