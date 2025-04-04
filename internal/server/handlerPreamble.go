package server

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type PreambleData struct {
	User          auth.AuthUser
	Offer         shared.Offer
	Quiz          shared.Quiz
	Languages     []shared.Language
	Participation shared.Participation
	Results       []Result
	Problems      []shared.Problem
}

type Result struct {
	Problem shared.Problem
	Score   shared.Score
}

type PreambleInput struct {
	QuizID string
}

type PreambleStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error)
	SelectProblems(ctx context.Context, quizID string) ([]shared.Problem, error)
	SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error)
	SelectOffer(ctx context.Context, id string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
}

func CreateParticipationHandler(templ TemplatesRepo, storage PreambleStorage, authService AuthRep, inputFn quizPageInputFunc) http.HandlerFunc {
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
				problems, err := storage.SelectProblems(r.Context(), quiz.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				results := make([]Result, 0)
				for _, problem := range problems {
					score, err := storage.SelectScore(r.Context(), user.ID, problem.ID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					result := Result{
						Problem: problem,
						Score: score,
					}
					results = append(results, result)
				}
				data.Results = results
				data.Participation = participation
			}
		}
		if err := templ.Render(w, "preamble", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (app *App) PreambleHandler() http.HandlerFunc {
	return CreateParticipationHandler(app.Templ, app.Storage, app.AuthService, GetQuizPageInput)
}
