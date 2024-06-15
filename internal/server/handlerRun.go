package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

type RunInput struct {
	ProblemID  string
	Src        string
	LanguageID int32
}

type RunOutput struct {
	SubmissionID string
}

func getRunInput(r *http.Request) (RunInput, error) {
	problemID := r.FormValue("problemID")
	src := r.FormValue("src")
	languageID := r.FormValue("languageID")
	if uuid.Validate(problemID) != nil {
		return RunInput{}, fmt.Errorf("problem_id is not a valid UUID")
	}
	if src == "" {
		return RunInput{}, fmt.Errorf("src is empty")
	}
	intLanguageID, err := strconv.Atoi(languageID)
	if err != nil {
		return RunInput{}, fmt.Errorf("languageID is not a valid integer")
	}
	if intLanguageID < 0 || intLanguageID > 100 {
		return RunInput{}, fmt.Errorf("languageID is not in the valid range")
	}
	input := RunInput{
		ProblemID:  problemID,
		Src:        src,
		LanguageID: int32(intLanguageID),
	}
	return input, nil
}

type RunStorage interface {
	//Also handle logic for participationID and quizz time
	CreateSubmission(ctx context.Context, submissionID, userID, problemID, src string, languageID int32) error

	GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error)
}

func createRunHandler(templ *Templates, storage RunStorage, authService AuthRep, st *codejudge.Stream, judge codejudge.Judge0, topicCleanup time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := getRunInput(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//Select DB Test Cases
		testCases, err := storage.GetTestCases(r.Context(), input.ProblemID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//DB INSERTS
		submissionID := uuid.NewString()
		err = storage.CreateSubmission(r.Context(), submissionID, userID, input.ProblemID, input.Src, input.LanguageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//Judge request
		tokens, err := judge.Send(testCases, submissionID, input.Src, input.LanguageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//STREAM
		err = st.Register(submissionID, tokens, topicCleanup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = templ.Render(w, "sseResults", RunOutput{SubmissionID: submissionID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (app *App) RunHandler() http.HandlerFunc {
	return createRunHandler(
    app.Templ,
		app.Storage,
		app.AuthService,
		app.Stream,
		app.Judge,
		60*time.Second,
	)
}
