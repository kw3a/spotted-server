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

type participateStorage struct{}

func (p *participateStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return nil
}

type invalidParticipateStorage struct{}

func (i *invalidParticipateStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return errors.New("error")
}

func participateInputFn(r *http.Request) (server.ParticipateInput, error) {
	return server.ParticipateInput{}, nil
}
func TestGetParticipateInputEmpty(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetParticipateInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetParticipateInputBadQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := server.GetParticipateInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetParticipateInput(t *testing.T) {
	quizID := uuid.NewString()
	formValues := map[string][]string{
		"quizID": {quizID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := server.GetParticipateInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestParticipateHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.ParticipateInput, error) {
		return server.ParticipateInput{}, errors.New("error")
	}
	handler := server.CreateParticipateHandler(&participateStorage{}, &authRepo{}, invalidInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req := formRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestParticipateHandlerBadAuth(t *testing.T) {
	handler := server.CreateParticipateHandler(&participateStorage{}, &invalidAuthRepo{}, participateInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req := formRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestParticipateHandlerBadStorage(t *testing.T) {
	handler := server.CreateParticipateHandler(&invalidParticipateStorage{}, &authRepo{}, participateInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req := formRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("expected bad request")
	}
}

func TestParticipateHandler(t *testing.T) {
	handler := server.CreateParticipateHandler(&participateStorage{}, &authRepo{}, participateInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req := formRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
}
