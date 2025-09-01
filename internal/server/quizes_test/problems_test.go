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

type problemsStorage struct{}

func (p *problemsStorage) SelectProblem(ctx context.Context, problemID string) (shared.Problem, error) {
	return shared.Problem{}, nil
}

type invalidProblemsStorage struct{}

func (i *invalidProblemsStorage) SelectProblem(ctx context.Context, problemID string) (shared.Problem, error) {
	return shared.Problem{}, errors.New("error")
}

func problemsInputFn(r *http.Request) (quizes.ProblemsInput, error) {
	return quizes.ProblemsInput{}, nil
}

func TestGetProblemsInputEmpty(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetProblemsInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetProblemsInputBadProblemID(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetProblemsInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetProblemsInput(t *testing.T) {
	problemID := uuid.NewString()
	formValues := map[string][]string{
		"problemID": {problemID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := quizes.GetProblemsInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.ProblemID != problemID {
		t.Error("invalid problem ID")
	}
}

func TestProblemsHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.ProblemsInput, error) {
		return quizes.ProblemsInput{}, errors.New("error")
	}
	handler := quizes.CreateProblemHandler(&templates{}, &problemsStorage{}, invalidInputFn)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/problems", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestProblemsHandlerBadStorage(t *testing.T) {
	handler := quizes.CreateProblemHandler(&templates{}, &invalidProblemsStorage{}, problemsInputFn)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/problems", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest{
		t.Error("expected bad request")
	}
}

func TestProblemsHandlerBadTemplate(t *testing.T) {
	handler := quizes.CreateProblemHandler(&invalidTemplates{}, &problemsStorage{}, problemsInputFn)
	if handler == nil {
		t.Error("expected nil")
	}
	req, _ := http.NewRequest("GET", "/problems", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected internal server error")
	}
}

func TestProblemsHandler(t *testing.T) {
	handler := quizes.CreateProblemHandler(&templates{}, &problemsStorage{}, problemsInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/problems", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
