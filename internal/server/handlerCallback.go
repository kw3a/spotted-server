package server

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/shopspring/decimal"
)

type CallbackURLParamsInput struct {
	SubmissionID string
	TestCaseID   string
}

func GetCallbackURLParamsInput(r *http.Request) (CallbackURLParamsInput, error) {
	submissionID := chi.URLParam(r, "submissionID")
	if uuid.Validate(submissionID) != nil {
		return CallbackURLParamsInput{}, errors.New("invalid submission ID")
	}
	testCaseID := chi.URLParam(r, "testCaseID")
	if uuid.Validate(testCaseID) != nil {
		return CallbackURLParamsInput{}, errors.New("invalid tc ID")
	}
	return CallbackURLParamsInput{
		SubmissionID: submissionID,
		TestCaseID:   testCaseID,
	}, nil
}

type CallbackJsonInput struct {
	Stdout        string          `json:"stdout"`
	Time          decimal.Decimal `json:"time"`
	Memory        int32           `json:"memory"`
	Stderr        string          `json:"stderr"`
	Token         string          `json:"token"`
	CompileOutput string          `json:"compile_output"`
	Message       string          `json:"message"`
	Status        status          `json:"status"`
}
type status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type CallbackStorage interface {
	UpdateTestCaseResult(ctx context.Context, input CallbackJsonInput, submissionID string, tcID string) error
}

type callbackInputFn func(r *http.Request) (CallbackURLParamsInput, error)
type decoderFn func(r *http.Request) (CallbackJsonInput, error)

func CreateCallbackHandler(storage CallbackStorage, st StreamService, decoder decoderFn, inputFn callbackInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParams, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		decoded, err := decoder(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(200)
		plainText, err := base64.StdEncoding.DecodeString(decoded.Stdout)
		if err != nil {
			log.Println(err)
			return
		}
		decoded.Stdout = string(plainText)
		err = storage.UpdateTestCaseResult(r.Context(), decoded, urlParams.SubmissionID, urlParams.TestCaseID)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Updated testCase, status: " + decoded.Status.Description)
		err = st.Update(urlParams.SubmissionID, decoded.Token, decoded.Status.Description)
		if err != nil {
			log.Printf("error updating stream: %v", err)
		}
	}
}

func (app *App) CallbackHandler() http.HandlerFunc {
	return CreateCallbackHandler(
		app.Storage,
		app.Stream,
		shared.Decode[CallbackJsonInput],
		GetCallbackURLParamsInput,
	)
}
