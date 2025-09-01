package quizestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/quizes"
)

func sourceInputFn(r *http.Request) (quizes.SourceInput, error) {
	return quizes.SourceInput{}, nil
}

type sourceStorage struct{}
func (s sourceStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	return "", nil
}

type invalidSourceStorage struct{}
func (i invalidSourceStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	return "", errors.New("error")
}

func TestGetSourceInputBadLanguageID(t *testing.T) {
	languageID := "60"
	formValues := map[string][]string{
		"problemID":  {"invalid"},
		"languageID": {languageID},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetSourceInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetSourceInputBadProblemID(t *testing.T) {
	problemID := uuid.NewString()
	formValues := map[string][]string{
		"problemID":  {problemID},
		"languageID": {"invalid"},
	}
	req := formRequest("GET", "/", formValues)
	_, err := quizes.GetSourceInput(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetSourceInput(t *testing.T) {
	problemID := uuid.NewString()
	languageID := "60"
	formValues := map[string][]string{
		"problemID":  {problemID},
		"languageID": {languageID},
	}
	req := formRequest("GET", "/", formValues)
	input, err := quizes.GetSourceInput(req)
	if err != nil {
		t.Error(err)
	}
	if input.ProblemID != problemID {
		t.Error("invalid problem ID")
	}
	intLanguageID := int(input.LanguageID)
	strLanguageID := strconv.Itoa(intLanguageID)
	if strLanguageID != languageID {
		t.Error("invalid language ID")
	}
}

func TestSourceHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.SourceInput, error) {
		return quizes.SourceInput{}, errors.New("error")
	}
	handler := quizes.CreateSourceHandler(sourceStorage{}, authRepo{}, invalidInputFn)
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

func TestSourceHandlerBadStorage(t *testing.T) {
	handler := quizes.CreateSourceHandler(invalidSourceStorage{}, authRepo{}, sourceInputFn)
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

func TestSourceHandlerBadAuth(t *testing.T) {
	handler := quizes.CreateSourceHandler(sourceStorage{}, invalidAuthRepo{}, sourceInputFn)
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

func TestSourceHandler(t *testing.T) {
	handler := quizes.CreateSourceHandler(sourceStorage{}, authRepo{}, sourceInputFn)
	if handler == nil {
		t.Error("expected handler")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Error("expected OK")
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Error("expected text/plain")
	}
}
