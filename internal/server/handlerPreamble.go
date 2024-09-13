package server

import (
	"context"
	"net/http"
	"time"
)

type PreambleData struct {
	Quiz             Quiz
	Participation    PreambleParticipation
	PreambleProblems []PreambleProblem
}

type PreambleProblem struct {
	Title             string
	NumberOfTestCases int
	AcceptedTestCases int
}

type PreambleParticipation struct {
	CreatedAt string
	ExpiresAt string
}

type PreambleInput struct {
	QuizID string
}

type Quiz struct {
	ID          string
	Title       string
	Description string
	Duration    int32
}

type ParticipationData struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type PreambleStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error)
	SelectProblemIDs(ctx context.Context, QuizID string) ([]string, error)
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
	SelectQuiz(ctx context.Context, id string) (Quiz, error)
}

func CreateParticipationHandler(templ TemplatesRepo, storage PreambleStorage, authService AuthRep, inputFn quizPageInputFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quiz, err := storage.SelectQuiz(r.Context(), input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := PreambleData{
			Quiz: quiz,
		}
		participation, err := storage.ParticipationStatus(r.Context(), userID, input.QuizID)
		if err == nil {
			problemIDs, err := storage.SelectProblemIDs(r.Context(), input.QuizID)
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
				score, err := storage.SelectScore(r.Context(), userID, problemID)
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
			data.Participation = PreambleParticipation{
				CreatedAt: participation.CreatedAt.Format("02-Jan-2006 15:04"),
				ExpiresAt: participation.ExpiresAt.Format("02-Jan-2006 15:04"),
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
