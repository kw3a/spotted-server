package quizes

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ParticipationStorage interface {
	Participate(ctx context.Context, userID string, quizID string) error
}

type ParticipateInput struct {
	QuizID string
}

func GetParticipateInput(r *http.Request) (ParticipateInput, error) {
	quizID := r.FormValue("quizID")
	if err := shared.ValidateUUID(quizID); err != nil {
		return ParticipateInput{}, err
	}
	return ParticipateInput{
		QuizID: quizID,
	}, nil
}
type participateInputFn func(r *http.Request) (ParticipateInput, error)
func CreateParticipateHandler(storage ParticipationStorage, authService shared.AuthRep, inputFn participateInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = storage.Participate(r.Context(), user.ID, input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("HX-Redirect", "/quizzes/"+input.QuizID)
		w.WriteHeader(http.StatusOK)
	}
}
