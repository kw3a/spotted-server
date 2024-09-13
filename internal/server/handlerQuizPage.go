package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type QuizPageData struct {
	QuizID         string
	Problems       []ProblemSelector
	ExpiresAt      time.Time
	Score          ScoreData
	ProblemContent ProblemContent
	Examples       []Example
	EditorData     EditorData
	Languages      []LanguageSelector
}

type ProblemSelector struct {
	ID          string
	ProblemName string
}

type ScoreData struct {
	AcceptedTestCases int
	TotalTestCases    int
}

type ProblemContent struct {
	Title       string
	Description string
}

type Example struct {
	Input  string
	Output string
}
type LanguageSelector struct {
	LanguageID    int32
	DisplayedName string
	SimpleName    string
}

type EditorData struct {
	SrcValue string
	Language string
}
type QuizPageInput struct {
	QuizID string
}

func GetQuizPageInput(r *http.Request) (QuizPageInput, error) {
	quizID := chi.URLParam(r, "quizID")
	if err := ValidateUUID(quizID); err != nil {
		return QuizPageInput{}, err
	}
	return QuizPageInput{
		QuizID: quizID,
	}, nil
}

type QuizPageStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error)
	SelectProblemIDs(ctx context.Context, quizID string) ([]string, error)
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
	SelectExamples(ctx context.Context, problemID string) ([]Example, error)
	SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error)
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
func SelectFirstLanguage(languages []LanguageSelector) LanguageSelector {
	return languages[0]
}
type quizPageInputFunc = func(r *http.Request) (QuizPageInput, error)
type enumProblemsFn = func([]string) []ProblemSelector
type selectProblemFn = func([]string) string
type selectLanguageFn = func([]LanguageSelector) LanguageSelector

func CreateQuizPageHandler(
	templ TemplatesRepo,
	storage QuizPageStorage,
	authRep AuthRep,
	redirectPath string,
	inputFn quizPageInputFunc,
	selectProblFn selectProblemFn,
	selectLangFn selectLanguageFn,
	enumerateProblemsFn enumProblemsFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authRep.GetUser(r)
		if err != nil {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		problemIDs, err := storage.SelectProblemIDs(r.Context(), input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		partiData, err := storage.ParticipationStatus(r.Context(), userID, input.QuizID)
		if err != nil || partiData.ExpiresAt.Before(time.Now()) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		selectedProblem := selectProblFn(problemIDs)
		score, err := storage.SelectScore(r.Context(), userID, selectedProblem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		problemContent, err := storage.SelectProblem(r.Context(), selectedProblem)
		if err != nil {
			http.Error(w, "problem description not found", http.StatusInternalServerError)
			return
		}
		examples, err := storage.SelectExamples(r.Context(), selectedProblem)
		if err != nil {
			http.Error(w, "error in find examples", http.StatusInternalServerError)
			return
		}
		languages, err := storage.SelectLanguages(r.Context(), input.QuizID)
		if err != nil {
			http.Error(w, "languages not found", http.StatusInternalServerError)
			return
		}
		selectedLanguage := selectLangFn(languages)
		lastSrc, err := storage.LastSrc(r.Context(), userID, selectedProblem, selectedLanguage.LanguageID)
		if err != nil {
			http.Error(w, "src not found", http.StatusInternalServerError)
			return
		}
		quizPageData := QuizPageData{
			QuizID:         input.QuizID,
			Problems:       enumerateProblemsFn(problemIDs),
			ExpiresAt:      partiData.ExpiresAt,
			Score:          score,
			ProblemContent: problemContent,
			Examples:       examples,
			EditorData:     EditorData{SrcValue: lastSrc, Language: selectedLanguage.SimpleName},
			Languages:      languages,
		}
		err = templ.Render(w, "quizPage", quizPageData)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't render quiz page: %s", err), http.StatusInternalServerError)
		}
	}
}

func (DI *App) QuizPageHandler() http.HandlerFunc {
	return CreateQuizPageHandler(
		DI.Templ,
		DI.Storage,
		DI.AuthService,
		"/",
		GetQuizPageInput,
		SelectFirstProblem,
		SelectFirstLanguage,
		EnumerateProblems,
	)
}
