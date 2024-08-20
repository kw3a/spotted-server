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

type problemsStorage struct{}

// SelectProblem implements server.ProblemsStorage.
func (p *problemsStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	return server.ProblemContent{}, nil
}

type invalidProblemsStorage struct{}

// SelectProblem implements server.ProblemsStorage.
func (i *invalidProblemsStorage) SelectProblem(ctx context.Context, problemID string) (server.ProblemContent, error) {
	return server.ProblemContent{}, errors.New("error")
}

func problemsInputFn(r *http.Request) (server.ProblemsInput, error) {
	return server.ProblemsInput{}, nil
}
func TestGetProblemsInputEmpty(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetProblemsInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetProblemsInputBadProblemID(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetProblemsInput(req)
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
	input, err := server.GetProblemsInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.ProblemID != problemID {
		t.Error("invalid problem ID")
	}
}

func TestProblemsHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.ProblemsInput, error) {
		return server.ProblemsInput{}, errors.New("error")
	}
	handler := server.CreateProblemHandler(&templates{}, &problemsStorage{}, invalidInputFn)
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
	handler := server.CreateProblemHandler(&templates{}, &invalidProblemsStorage{}, problemsInputFn)
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
	handler := server.CreateProblemHandler(&invalidTemplates{}, &problemsStorage{}, problemsInputFn)
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
	handler := server.CreateProblemHandler(&templates{}, &problemsStorage{}, problemsInputFn)
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
