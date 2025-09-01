package quizestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/quizes"
)

type participateStorage struct{}

func (p *participateStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return nil
}

type invalidParticipateStorage struct{}

func (i *invalidParticipateStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return errors.New("error")
}

func participateInputFn(r *http.Request) (quizes.ParticipateInput, error) {
	return quizes.ParticipateInput{QuizID: "1"}, nil
}
func TestGetParticipateInputEmpty(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {""},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetParticipateInput(req)
	if err == nil {
		t.Error("expected error")
	}
}
func TestGetParticipateInputBadQuizID(t *testing.T) {
	formValues := map[string][]string{
		"quizID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetParticipateInput(req)
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
	input, err := quizes.GetParticipateInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestParticipateHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.ParticipateInput, error) {
		return quizes.ParticipateInput{}, errors.New("error")
	}
	handler := quizes.CreateParticipateHandler(&participateStorage{}, &authRepo{}, invalidInputFn)
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
	handler := quizes.CreateParticipateHandler(&participateStorage{}, &invalidAuthRepo{}, participateInputFn)
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
	handler := quizes.CreateParticipateHandler(&invalidParticipateStorage{}, &authRepo{}, participateInputFn)
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
	handler := quizes.CreateParticipateHandler(&participateStorage{}, &authRepo{}, participateInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req := formRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected ok")
	}
	if w.Header().Get("HX-Redirect") != "/quizzes/1" {
		t.Error("invalid redirect")
	}
}
