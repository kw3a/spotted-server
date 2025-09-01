package quizestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type endStorage struct{}

func (e endStorage) EndQuiz(ctx context.Context, userID string, quizID string) (shared.Offer, error) {
	return shared.Offer{ID: "1"}, nil
}

type invalidEndStorage struct{}

func (i invalidEndStorage) EndQuiz(ctx context.Context, userID string, quizID string) (shared.Offer, error) {
	return shared.Offer{}, errors.New("error")
}

func endInputFn(r *http.Request) (quizes.EndInput, error) {
	return quizes.EndInput{QuizID: "1"}, nil
}

func TestGetEndInputEmptyQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetEndInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetEndInputInvalidQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetEndInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetEndInput(t *testing.T) {
	quizID := uuid.NewString()
	formValues := map[string][]string{
		"quizID": {quizID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := quizes.GetEndInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestEndHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.EndInput, error) {
		return quizes.EndInput{}, errors.New("error")
	}
	handler := quizes.CreateEndHandler(&endStorage{}, &authRepo{}, invalidInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("invalid status code")
	}
}

func TestEndHandlerInvalidAuth(t *testing.T) {
	handler := quizes.CreateEndHandler(&endStorage{}, &invalidAuthRepo{}, endInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Error("invalid status code")
	}
}

func TestEndHandlerInvalidStorage(t *testing.T) {
	handler := quizes.CreateEndHandler(&invalidEndStorage{}, &authRepo{}, endInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("invalid status code")
	}
}

func TestEndHandler(t *testing.T) {
	handler := quizes.CreateEndHandler(&endStorage{}, &authRepo{}, endInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("invalid status code")
	}
	if w.Header().Get("HX-Redirect") != "/preamble/1" {
		t.Errorf("invalid redirect. want: '/preamble/1', got:%s",w.Header().Get("HX-Redirect"))
	}
}
