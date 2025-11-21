package offers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type PreambleData struct {
	User          auth.AuthUser
	Offer         shared.Offer
	Quiz          shared.Quiz
	Languages     []shared.Language
	Participation shared.Participation
	QuizAlive     bool
}

type Result struct {
	Problem shared.Problem
	Score   shared.Score
}

type PreambleInput struct {
	OfferID string
}

func GetPreambleInput(r *http.Request) (PreambleInput, error) {
	quizID := chi.URLParam(r, "quizID")
	if err := shared.ValidateUUID(quizID); err != nil {
		return PreambleInput{}, err
	}
	return PreambleInput{
		OfferID: quizID,
	}, nil
}

type PreambleStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error)
	SelectOffer(ctx context.Context, id string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
}

type preambleInputFunc = func(r *http.Request) (PreambleInput, error)

func CreateParticipationHandler(templ shared.TemplatesRepo, storage PreambleStorage, authService shared.AuthRep, inputFn preambleInputFunc) http.HandlerFunc {
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
		offer, err := storage.SelectOffer(r.Context(), input.OfferID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data := PreambleData{
			Offer: offer,
			User:  user,
		}
		if offer.Status != 0 {
			quiz, err := storage.SelectQuizByOffer(r.Context(), offer.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			languages, err := storage.SelectLanguages(r.Context(), quiz.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.Quiz = quiz
			data.Languages = languages
			participation, err := storage.ParticipationStatus(r.Context(), user.ID, quiz.ID)
			if err == nil {
				data.Participation = participation
				if data.Participation.ExpiresAt.After(time.Now()) {
					data.QuizAlive = true
				}
			}
		}
		if err := templ.Render(w, "preamble", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
