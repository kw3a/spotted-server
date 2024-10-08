package server

import (
	"context"
	"net/http"
)

type EndStorage interface {
	EndQuiz(ctx context.Context, userID, quizID string) error
}
type EndInput struct {
	QuizID string
}

func GetEndInput(r *http.Request) (EndInput, error) {
	quizID := r.FormValue("quizID")
	if err := ValidateUUID(quizID); err != nil {
		return EndInput{}, err
	}
	return EndInput{
		QuizID: quizID,
	}, nil
}
type endInputFn func(r *http.Request) (EndInput, error)
func CreateEndHandler(endStorage EndStorage, authService AuthRep, inputFn endInputFn) http.HandlerFunc {
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
		err = endStorage.EndQuiz(r.Context(), user, input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("HX-Redirect", "/preamble/"+input.QuizID)
		w.WriteHeader(http.StatusOK)
	}
}

func (DI *App) EndHandler() http.HandlerFunc {
	return CreateEndHandler(DI.Storage, DI.AuthService, GetEndInput)
}
