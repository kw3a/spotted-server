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

func scoreInputFn(r *http.Request) (server.ScoreInput, error) {
	return server.ScoreInput{}, nil
}

type scoreStorage struct{}

func (s scoreStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	return server.ScoreData{}, nil
}

type invalidScoreStorage struct{}
func (i *invalidScoreStorage) SelectScore(ctx context.Context, userID string, problemID string) (server.ScoreData, error) {
	return server.ScoreData{}, errors.New("error")
}

func TestGetScoreInputEmpty(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetScoreInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetScoreInputBadQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetScoreInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetScoreInput(t *testing.T) {
	problemID := uuid.NewString()
	formValues := map[string][]string{
		"problemID": {problemID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := server.GetScoreInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.ProblemID != problemID {
		t.Error("invalid problem ID")
	}
}

func TestScoreHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.ScoreInput, error) {
		return server.ScoreInput{}, errors.New("error")
	}
	handler := server.CreateScoreHandler(&templates{}, scoreStorage{}, authRepo{}, invalidInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestScoreHandlerBadStorage(t *testing.T) {
	handler := server.CreateScoreHandler(&templates{}, &invalidScoreStorage{}, authRepo{}, scoreInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}
func TestScoreHandlerBadAuth(t *testing.T) {
	handler := server.CreateScoreHandler(&templates{}, scoreStorage{}, &invalidAuthRepo{}, scoreInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Error("expected unauthorized")
	}
}

func TestScoreHandlerBadTemplate(t *testing.T) {
	handler := server.CreateScoreHandler(&invalidTemplates{}, scoreStorage{}, authRepo{}, scoreInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Error("expected internal server error")
	}
}

func TestScoreHandler(t *testing.T) {
	handler := server.CreateScoreHandler(&templates{}, scoreStorage{}, authRepo{}, scoreInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
