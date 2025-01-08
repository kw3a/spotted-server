package server

import (
	"context"
	"net/http"
	"time"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type PreambleData struct {
	User             auth.AuthUser
	Offer            shared.Offer
	Quiz             shared.Quiz
	Languages        []shared.Language
	Participation    ParticipationData
	PreambleProblems []PreambleProblem
}

type PreambleProblem struct {
	Title             string
	NumberOfTestCases int
	AcceptedTestCases int
}

type PreambleInput struct {
	QuizID string
}

type ParticipationData struct {
	ID           string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RelativeTime string
}

type PreambleStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error)
	SelectProblemIDs(ctx context.Context, QuizID string) ([]string, error)
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
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
				problemIDs, err := storage.SelectProblemIDs(r.Context(), quiz.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				preambleProblems := make([]PreambleProblem, 0)
				for _, problemID := range problemIDs {
					problemContent, err := storage.SelectProblem(r.Context(), problemID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					score, err := storage.SelectScore(r.Context(), user.ID, problemID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					problem := PreambleProblem{
						Title:             problemContent.Title,
						NumberOfTestCases: score.TotalTestCases,
						AcceptedTestCases: score.AcceptedTestCases,
					}
					preambleProblems = append(preambleProblems, problem)
				}
				data.PreambleProblems = preambleProblems
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
