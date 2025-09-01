package quizestest

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type callbackStorageMock struct {
	mock.Mock
}

func (storage *callbackStorageMock) UpdateTestCaseResult(
	ctx context.Context,
	input shared.CallbackJsonInput,
	submissionID string,
	tcID string) error {
	args := storage.Called(ctx, input, submissionID, tcID)
	return args.Error(0)
}

func callbackInputFn(r *http.Request) (quizes.CallbackURLParamsInput, error) {
	return quizes.CallbackURLParamsInput{}, nil
}

func jsonDecoder(r *http.Request) (shared.CallbackJsonInput, error) {
	return shared.CallbackJsonInput{}, nil
}

func invalidJsonDecoder(r *http.Request) (shared.CallbackJsonInput, error) {
	return shared.CallbackJsonInput{}, fmt.Errorf("error")
}

func TestCallbackUrlParamsEmpty(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	_, err = quizes.GetCallbackURLParamsInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallbackUrlParamsInvalidSubmissionID(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": "invalid",
		"testCaseID":   uuid.NewString(),
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	_, err = quizes.GetCallbackURLParamsInput(reqWithUrlParam)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallbackUrlParamsInvalidTestCaseID(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": uuid.NewString(),
		"testCaseID":   "invalid",
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	_, err = quizes.GetCallbackURLParamsInput(reqWithUrlParam)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallbackUrlParams(t *testing.T) {
	submissionID := uuid.NewString()
	tcID := uuid.NewString()
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": submissionID,
		"testCaseID":   tcID,
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	params, err := quizes.GetCallbackURLParamsInput(reqWithUrlParam)
	if err != nil {
		t.Error(err)
	}
	if params.SubmissionID != submissionID {
		t.Error("invalid submission ID")
	}
	if params.TestCaseID != tcID {
		t.Error("invalid tc ID")
	}
}

func TestCallbackHandlerBadURLInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.CallbackURLParamsInput, error) {
		return quizes.CallbackURLParamsInput{}, fmt.Errorf("error")
	}
	handler := quizes.CreateCallbackHandler(
		&callbackStorageMock{},
		&streamService{},
		nil,
		invalidInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCallbackHandlerBadJsonInput(t *testing.T) {
	handler := quizes.CreateCallbackHandler(
		&callbackStorageMock{},
		&streamService{},
		invalidJsonDecoder,
		callbackInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCallbackHandlerBadStorage(t *testing.T) {
	storage := new(callbackStorageMock)
	storage.On("UpdateTestCaseResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage error"))
	handler := quizes.CreateCallbackHandler(
		storage,
		&streamService{},
		jsonDecoder,
		callbackInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, logBuf.String(), "storage error")
}

func TestCallbackHandlerBadStream(t *testing.T) {
	storage := new(callbackStorageMock)
	storage.On("UpdateTestCaseResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	stream := new(streamService)
	stream.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("stream error"))
	handler := quizes.CreateCallbackHandler(
		storage,
		stream,
		jsonDecoder,
		callbackInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, logBuf.String(), "error updating stream:")
}

func TestCallbackHandler(t *testing.T) {
	storage := new(callbackStorageMock)
	storage.On("UpdateTestCaseResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	stream := new(streamService)
	stream.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := quizes.CreateCallbackHandler(
		storage,
		stream,
		jsonDecoder,
		callbackInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.NotContains(t, logBuf.String(), "error updating stream:")
}
