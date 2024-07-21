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
	QuizID     string
	ProblemID  string
	Src        string
	LanguageID int32
}

type RunOutput struct {
	SubmissionID string
}

func getRunInput(r *http.Request) (RunInput, error) {
  quizID := r.FormValue("quizID")
	problemID := r.FormValue("problemID")
	src := r.FormValue("src")
	languageID := r.FormValue("languageID")
  if uuid.Validate(quizID) != nil {
    return RunInput{}, fmt.Errorf("quiz_id is not a valid UUID")
  }
	if uuid.Validate(problemID) != nil {
		return RunInput{}, fmt.Errorf("problem_id is not a valid UUID")
	}
	if src == "" {
		return RunInput{}, fmt.Errorf("src is empty")
	}
	languageIDInt, err := strconv.ParseInt(languageID, 10, 32)
	if err != nil {
		return RunInput{}, fmt.Errorf("languageID is not a valid integer")
	}
	languageIDInt32 := int32(languageIDInt)
	if languageIDInt32 < 0 || languageIDInt32 > 100 {
		return RunInput{}, fmt.Errorf("languageID is not in the valid range")
	}
	input := RunInput{
    QuizID:     quizID,
		ProblemID:  problemID,
		Src:        src,
		LanguageID: languageIDInt32,
	}
	return input, nil
}

type RunStorage interface {
	//Also handle logic for participationID and quizz time
	CreateSubmission(ctx context.Context, submissionID, participationID, problemID, src string, languageID int32) error

	ParticipationStatus(ctx context.Context, userID string, quizID string) (string, bool, time.Time, error)
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
		participationID, inHour, _, err := storage.ParticipationStatus(r.Context(), userID, input.QuizID)
		if err != nil || !inHour {
			http.Error(w, "error in getting status:"+err.Error(), http.StatusBadRequest)
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
		err = storage.CreateSubmission(r.Context(), submissionID, participationID, input.ProblemID, input.Src, input.LanguageID)
		if err != nil {
			http.Error(w, "error in create submission: "+err.Error(), http.StatusInternalServerError)
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
