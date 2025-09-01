package quizes

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
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

type CallbackStorage interface {
	UpdateTestCaseResult(ctx context.Context, input shared.CallbackJsonInput, submissionID string, tcID string) error
}

type callbackInputFn func(r *http.Request) (CallbackURLParamsInput, error)
type decoderFn func(r *http.Request) (shared.CallbackJsonInput, error)

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
