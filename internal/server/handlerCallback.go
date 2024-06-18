package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/utils"
	"github.com/shopspring/decimal"
)


type CallbackUrlParams struct {
	SubmissionID string
	TestCaseID   string
}

func NewCallbackUrlParams(r *http.Request) (CallbackUrlParams, error) {
	submissionID := chi.URLParam(r, "submissionID")
	if uuid.Validate(submissionID) != nil {
		return CallbackUrlParams{}, errors.New("invalid submission ID")
	}

	testCaseID := chi.URLParam(r, "testCaseID")
	if uuid.Validate(testCaseID) != nil {
		return CallbackUrlParams{}, errors.New("invalid tc ID")
	}
	return CallbackUrlParams{
		SubmissionID: submissionID,
		TestCaseID:   testCaseID,
	}, nil
}

type CallbackInput struct {
	Stdout        interface{}     `json:"stdout"`
	Time          decimal.Decimal `json:"time"`
	Memory        int32             `json:"memory"`
	Stderr        string          `json:"stderr"`
	Token         string          `json:"token"`
	CompileOutput interface{}     `json:"compile_output"`
	Message       string          `json:"message"`
	Status        status          `json:"status"`
}
type status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

func (input CallbackInput) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	return problems
}

type CallbackStorage interface {
  UpdateTestCaseResult(ctx context.Context, input CallbackInput, submissionID string, tcID string) error
}

func createCallbackHandler(storage CallbackStorage, st *codejudge.Stream) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParams, err := NewCallbackUrlParams(r)
		if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		decoded, problems, err := utils.DecodeValid[CallbackInput](r)
		if err != nil {
      http.Error(w, fmt.Sprintf("problems:\n%v", problems), http.StatusBadRequest)
			return
		}
		w.WriteHeader(200)

		err = storage.UpdateTestCaseResult(r.Context(), decoded, urlParams.SubmissionID, urlParams.TestCaseID)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Updated testCase, status: " + decoded.Status.Description)
		topic, err := st.GetTopic(urlParams.SubmissionID)
		if err != nil {
			log.Println(err)
			return
		}
		err = topic.Update(decoded.Token, decoded.Status.Description)
    if err != nil {
      log.Println(err)
    }
	}
}

func (app *App) CallbackHandler() http.HandlerFunc {
  return createCallbackHandler(
    app.Storage,
    app.Stream,
  )
}
