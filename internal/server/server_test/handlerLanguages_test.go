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

type languageStorage struct{}
func (l *languageStorage) SelectLanguages(ctx context.Context, quizID string) ([]server.LanguageSelector, error) {
	return []server.LanguageSelector{}, nil
}

type invalidLanguageStorage struct{}
func (i *invalidLanguageStorage) SelectLanguages(ctx context.Context, quizID string) ([]server.LanguageSelector, error) {
	return nil, errors.New("error")
}

func languagesInputFn(r *http.Request) (server.LanguagesInput, error) {
	return server.LanguagesInput{}, nil
}

func TestGetLanguagesInputEmpty(t *testing.T) {
	params := map[string]string{
		"quizID": "",
	}
	req, _ := http.NewRequest("GET", "/example", nil)
	reqWithUrlParam := WithUrlParams(req, params)
	_, err := server.GetLanguagesInput(reqWithUrlParam)
	if err == nil {
		t.Error(err)
	}
}

func TestGetLanguagesInputBadProblemID(t *testing.T) {
	params := map[string]string{
		"quizID": "invalid",
	}
	req, _ := http.NewRequest("GET", "/example", nil)
	reqWithUrlParam := WithUrlParams(req, params)
	_, err := server.GetLanguagesInput(reqWithUrlParam)
	if err == nil {
		t.Error(err)
	}
}
func TestGetLanguagesInput(t *testing.T) {
	quizID := uuid.NewString()
	params := map[string]string{
		"quizID": quizID,
	}
	req, _ := http.NewRequest("GET", "/example", nil)
	reqWithUrlParam := WithUrlParams(req, params)
	input, err := server.GetLanguagesInput(reqWithUrlParam)
	if err != nil {
		t.Error(err)
	}
	if input.QuizID != quizID {
		t.Error("invalid quiz ID")
	}
}

func TestLanguagesHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (server.LanguagesInput, error) {
		return server.LanguagesInput{}, errors.New("error")
	}
	handler := server.CreateLanguagesHandler(&templates{}, &languageStorage{}, invalidInputFn)
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

func TestLanguagesHandlerBadStorage(t *testing.T) {
	handler := server.CreateLanguagesHandler(&templates{}, &invalidLanguageStorage{}, languagesInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest{
		t.Error("expected bad request")
	}
}

func TestLanguagesHandlerBadTemplates(t *testing.T) {
	handler := server.CreateLanguagesHandler(&invalidTemplates{}, &languageStorage{}, languagesInputFn)
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

func TestLanguagesHandler(t *testing.T) {
	handler := server.CreateLanguagesHandler(&templates{}, &languageStorage{}, languagesInputFn)
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
