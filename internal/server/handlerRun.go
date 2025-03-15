package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/shared"
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

func GetRunInput(r *http.Request) (RunInput, error) {
	quizID := r.FormValue("quizID")
	if err := ValidateUUID(quizID); err != nil {
		return RunInput{}, fmt.Errorf("quiz_id: %s", err.Error())
	}
	problemID := r.FormValue("problemID")
	if err := ValidateUUID(problemID); err != nil {
		return RunInput{}, fmt.Errorf("problem_id: %s", err.Error())
	}
	src := r.FormValue("src")
	if src == "" {
		return RunInput{}, fmt.Errorf("src is empty")
	}
	languageID := r.FormValue("languageID")
	languageIDInt32, err := shared.ValidateLanguageID(languageID)
	if err != nil {
		return RunInput{}, err
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

	ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error)
	GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error)
}
type JudgeService interface {
	Send(dbTestCases []codejudge.TestCase, submission codejudge.Submission) ([]string, error)
}

type StreamService interface {
	Register(name string, tokens []string, duration time.Duration) error
	Listen(name string) (chan string, error)
	Update(name, token, status string) error
}

type runInputFn func(r *http.Request) (RunInput, error)

func CreateRunHandler(
	templ TemplatesRepo,
	storage RunStorage,
	authService AuthRep,
	st StreamService,
	judge JudgeService,
	topicCleanup time.Duration,
	inputFn runInputFn,
) http.HandlerFunc {
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
		participation, err := storage.ParticipationStatus(r.Context(), user.ID, input.QuizID)
		if err != nil {
			http.Error(w, "error in getting status:"+err.Error(), http.StatusBadRequest)
			return
		}
		if participation.ExpiresAt.Before(time.Now()) {
			http.Error(w, "your participation is over", http.StatusUnauthorized)
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
		err = storage.CreateSubmission(r.Context(), submissionID, participation.ID, input.ProblemID, input.Src, input.LanguageID)
		if err != nil {
			http.Error(w, "error in create submission: "+err.Error(), http.StatusInternalServerError)
			return
		}
		//Judge request
		submission := codejudge.Submission{
			ID:         submissionID,
			Src:        input.Src,
			LanguageID: input.LanguageID,
		}
		tokens, err := judge.Send(testCases, submission)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	return CreateRunHandler(
		app.Templ,
		app.Storage,
		app.AuthService,
		app.Stream,
		&app.Judge,
		60*time.Second,
		GetRunInput,
	)
}
