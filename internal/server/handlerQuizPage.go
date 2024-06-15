package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type QuizPageData struct {
	QuizID         string
	Problems       []ProblemSelector
	ProblemContent ProblemContent
	Examples       []Example
	EditorData     EditorData
	Languages      []LanguageSelector
}

type ProblemSelector struct {
	ID          string
	ProblemName string
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

type QuizPageStorage interface {
	SelectProblemIDs(ctx context.Context, quizID string) ([]string, error)
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
	SelectExamples(ctx context.Context, problemID string) ([]Example, error)
	SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error)
	LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error)
}

func createQuizPageHandler(templ *Templates, storage QuizPageStorage, authRep AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizID := chi.URLParam(r, "quizID")
		if quizID == "" {
			http.Error(w, "invalid quiz id", http.StatusBadRequest)
			return
		}
		userID, err := authRep.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		problemIDs, err := storage.SelectProblemIDs(r.Context(), quizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		problems := []ProblemSelector{}
		for i, problemID := range problemIDs {
			strName := strconv.Itoa(i + 1)
			current := ProblemSelector{
				ID:          problemID,
				ProblemName: strName,
			}
			problems = append(problems, current)
		}
		selectedProblem := problemIDs[0]
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
		languages, err := storage.SelectLanguages(r.Context(), quizID)
		if err != nil {
			http.Error(w, "languages not found", http.StatusInternalServerError)
			return
		}
		selectedLanguage := languages[0]
		lastSrc, err := storage.LastSrc(r.Context(), userID, selectedProblem, selectedLanguage.LanguageID)
		if err != nil {
			http.Error(w, "src not found", http.StatusInternalServerError)
			return
		}
		quizPageData := QuizPageData{
			QuizID:         quizID,
			Problems:       problems,
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
	return createQuizPageHandler(
    DI.Templ,
		DI.Storage,
		DI.AuthService,
	)
}
