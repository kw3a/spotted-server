package server

import (
	"context"
	"database/sql"
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

type QuizPageStorage interface {
	ParticipationStatus(ctx context.Context, userID string, quizID string) (string, bool, time.Time, error)
	Participate(ctx context.Context, userID string, quizID string) error
	SelectProblemIDs(ctx context.Context, quizID string) ([]string, error)
	SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error)
	SelectProblem(ctx context.Context, problemID string) (ProblemContent, error)
	SelectExamples(ctx context.Context, problemID string) ([]Example, error)
	SelectLanguages(ctx context.Context, quizID string) ([]LanguageSelector, error)
	LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error)
}

func Participation(ctx context.Context, storage QuizPageStorage, userID, quizID string) (time.Time, error) {
	_, inHour, expiresAt, err := storage.ParticipationStatus(ctx, userID, quizID)
	if err != nil {
		if err != sql.ErrNoRows {
			return time.Now(), fmt.Errorf("error different from sql.ErrNoRows %s", err.Error()) 
		}
		err = storage.Participate(ctx, userID, quizID)
		if err != nil {
			return time.Now(), fmt.Errorf("error in participate %s", err.Error()) 
		}
		return time.Now(), nil
	}
	if !inHour {
		return time.Now(), fmt.Errorf("your participation is over") 
	}
	return expiresAt, nil 
}

func createQuizPageHandler(templ *Templates, storage QuizPageStorage, authRep AuthRep, redirectPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quizID := chi.URLParam(r, "quizID")
		if quizID == "" {
			http.Error(w, "invalid quiz id", http.StatusBadRequest)
			return
		}
		userID, err := authRep.GetUser(r)
		if err != nil {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}
		problemIDs, err := storage.SelectProblemIDs(r.Context(), quizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		expiresAt, err := Participation(r.Context(), storage, userID, quizID)
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
			ExpiresAt:      expiresAt,
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
	return createQuizPageHandler(
		DI.Templ,
		DI.Storage,
		DI.AuthService,
		"/",
	)
}
