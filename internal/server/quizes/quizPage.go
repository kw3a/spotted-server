package quizes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type QuizPageData struct {
	QuizID          string
	Problems        []ProblemSelector
	ExpiresAt       time.Time
	ParticipationID string
	Score           shared.Score
	Problem         shared.Problem
	Examples        []shared.Example
	EditorData      EditorData
	Languages       []shared.Language
}

type ProblemSelector struct {
	ID          string
	ProblemName string
}

type EditorData struct {
	SrcValue string
	Language string
}
type QuizPageInput struct {
	OfferID string
}

func GetQuizPageInput(r *http.Request) (QuizPageInput, error) {
	quizID := chi.URLParam(r, "quizID")
	if err := shared.ValidateUUID(quizID); err != nil {
		return QuizPageInput{}, err
	}
	return QuizPageInput{
		OfferID: quizID,
	}, nil
}

type QuizPageStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error)
	SelectProblemIDs(ctx context.Context, quizID string) ([]string, error)
	SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error)
	SelectProblem(ctx context.Context, problemID string) (shared.Problem, error)
	SelectExamples(ctx context.Context, problemID string) ([]shared.Example, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
	LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error)
}

func EnumerateProblems(problemIDs []string) []ProblemSelector {
	problems := []ProblemSelector{}
	for i, problemID := range problemIDs {
		strName := strconv.Itoa(i + 1)
		current := ProblemSelector{
			ID:          problemID,
			ProblemName: strName,
		}
		problems = append(problems, current)
	}
	return problems
}
func SelectFirstProblem(problemIDs []string) string {
	return problemIDs[0]
}
func SelectFirstLanguage(languages []shared.Language) shared.Language {
	return languages[0]
}

type quizPageInputFunc = func(r *http.Request) (QuizPageInput, error)
type enumProblemsFn = func([]string) []ProblemSelector
type selectProblemFn = func([]string) string
type selectLanguageFn = func([]shared.Language) shared.Language

func CreateQuizPageHandler(
	templ shared.TemplatesRepo,
	storage QuizPageStorage,
	authRep shared.AuthRep,
	inputFn quizPageInputFunc,
	selectProblFn selectProblemFn,
	selectLangFn selectLanguageFn,
	enumerateProblemsFn enumProblemsFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authRep.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		partiData, err := storage.ParticipationStatus(r.Context(), user.ID, input.OfferID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if partiData.ExpiresAt.Before(time.Now()) {
			http.Error(w, "your participation is over", http.StatusUnauthorized)
			return
		}
		problemIDs, err := storage.SelectProblemIDs(r.Context(), input.OfferID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		selectedProblem := selectProblFn(problemIDs)
		score, err := storage.SelectScore(r.Context(), user.ID, selectedProblem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		problem, err := storage.SelectProblem(r.Context(), selectedProblem)
		if err != nil {
			http.Error(w, "problem description not found", http.StatusInternalServerError)
			return
		}
		examples, err := storage.SelectExamples(r.Context(), selectedProblem)
		if err != nil {
			http.Error(w, "error in find examples", http.StatusInternalServerError)
			return
		}
		languages, err := storage.SelectLanguages(r.Context(), input.OfferID)
		if err != nil {
			http.Error(w, "languages not found", http.StatusInternalServerError)
			return
		}
		selectedLanguage := selectLangFn(languages)
		lastSrc, err := storage.LastSrc(r.Context(), user.ID, selectedProblem, selectedLanguage.ID)
		if err != nil {
			http.Error(w, "src not found", http.StatusInternalServerError)
			return
		}
		data := QuizPageData{
			QuizID:          input.OfferID,
			Problems:        enumerateProblemsFn(problemIDs),
			ExpiresAt:       partiData.ExpiresAt,
			ParticipationID: partiData.ID,
			Score:           score,
			Problem:         problem,
			Examples:        examples,
			EditorData:      EditorData{SrcValue: lastSrc, Language: selectedLanguage.Name},
			Languages:       languages,
		}
		err = templ.Render(w, "quizPage", data)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't render quiz page: %s", err), http.StatusInternalServerError)
		}
	}
}
