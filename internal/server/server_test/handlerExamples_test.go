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


type examplesStorage struct{}

func (e examplesStorage) SelectExamples(ctx context.Context, problemID string) ([]server.Example, error) {
	return nil, nil
}

type invalidExamplesStorage struct{}

func (i invalidExamplesStorage) SelectExamples(ctx context.Context, problemID string) ([]server.Example, error) {
	return nil, errors.New("error")
}

func examplesInputFn(r *http.Request) (server.ExamplesInput, error) {
	return server.ExamplesInput{}, nil
}

func TestGetExamplesInputEmptyProblemID(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetExamplesInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetExamplesInputInvalidProblemID(t *testing.T) {
	formValues := map[string][]string{
		"problemID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetExamplesInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetExamplesInput(t *testing.T) {
	problemID := uuid.NewString()
	formValues := map[string][]string{
		"problemID": {problemID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := server.GetExamplesInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.ProblemID != problemID {
		t.Error("invalid problem ID")
	}
}

func TestExamplesHandlerBadInput(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	invalidInputFn := func(r *http.Request) (server.ExamplesInput, error) {
		return server.ExamplesInput{}, errors.New("error")
	}
	handler := server.CreateExamplesHandler(&templates{}, examplesStorage{}, invalidInputFn)
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected 400")
	}
}

func TestExamplesHandlerBadStorage(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := server.CreateExamplesHandler(&templates{}, invalidExamplesStorage{}, examplesInputFn)
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected 400")
	}
}

func TestExamplesHandlerBadTemplate(t *testing.T) {
	problemID := uuid.NewString()
	formValues := map[string][]string{
		"problemID": {problemID},
	}
	req := formRequest("GET", "/", formValues)
	w := httptest.NewRecorder()
	handler := server.CreateExamplesHandler(&invalidTemplates{}, examplesStorage{}, examplesInputFn)
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected 500")
	}
}

func TestExamplesHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := server.CreateExamplesHandler(&templates{}, examplesStorage{}, examplesInputFn)
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected 200")
	}
}
