package servertest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
)

type endStorage struct{}

func (e endStorage) EndQuiz(ctx context.Context, userID string, quizID string) error {
	return nil
}

type invalidEndStorage struct{}

func (i invalidEndStorage) EndQuiz(ctx context.Context, userID string, quizID string) error {
	return errors.New("error")
}

func endInputFn(r *http.Request) (server.EndInput, error) {
	return server.EndInput{}, nil
}

func TestGetEndInputEmptyQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {""},
	}
	req := formRequest("GET", "/", formValues)

	_, err := server.GetEndInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetEndInputInvalidQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetEndInput(req)
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
	input, err := server.GetEndInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestEndHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.EndInput, error) {
		return server.EndInput{}, errors.New("error")
	}
	handler := server.CreateEndHandler(&endStorage{}, &authRepo{}, invalidInputFn)
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
	handler := server.CreateEndHandler(&endStorage{}, &invalidAuthRepo{}, endInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}

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

func TestEndHandlerInvalidStorage(t *testing.T) {
	handler := server.CreateEndHandler(&invalidEndStorage{}, &authRepo{}, endInputFn)
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
	handler := server.CreateEndHandler(&endStorage{}, &authRepo{}, endInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("invalid status code")
	}
}
