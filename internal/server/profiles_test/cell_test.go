package profilestest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/profiles"
	"github.com/stretchr/testify/mock"
)

type cellphoneStorage struct {
	mock.Mock
}

func (s *cellphoneStorage) UpdateCell(ctx context.Context, userID string, cell string) error {
	args := s.Called(ctx, userID, cell)
	return args.Error(0)
}

func cellphoneInputFn(r *http.Request) (profiles.CellUpdateInput, profiles.CellUpdateErrors, bool) {
	return profiles.CellUpdateInput{}, profiles.CellUpdateErrors{}, false
}

func TestUpdateCellHandlerBadAuth(t *testing.T) {
	storage := new(cellphoneStorage)
	handler := profiles.CreateUpdateCellHandler(&templates{}, invalidAuthRepo{}, storage, cellphoneInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateCellHandlerBadInputT(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.CellUpdateInput, profiles.CellUpdateErrors, bool) {
		return profiles.CellUpdateInput{}, profiles.CellUpdateErrors{}, true
	}
	storage := new(cellphoneStorage)
	handler := profiles.CreateUpdateCellHandler(&invalidTemplates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUpdateCellHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (profiles.CellUpdateInput, profiles.CellUpdateErrors, bool) {
		return profiles.CellUpdateInput{}, profiles.CellUpdateErrors{}, true
	}
	storage := new(cellphoneStorage)
	handler := profiles.CreateUpdateCellHandler(&templates{}, &authRepo{}, storage, invalidInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateCellHandlerBadStorageT(t *testing.T) {
	storage := new(cellphoneStorage)
	storage.On("UpdateCell", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	handler := profiles.CreateUpdateCellHandler(&invalidTemplates{}, &authRepo{}, storage, cellphoneInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUpdateCellHandlerBadStorage(t *testing.T) {
	storage := new(cellphoneStorage)
	storage.On("UpdateCell", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	handler := profiles.CreateUpdateCellHandler(&templates{}, &authRepo{}, storage, cellphoneInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateCellHandlerBadTemplate(t *testing.T) {
	storage := new(cellphoneStorage)
	storage.On("UpdateCell", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateUpdateCellHandler(&invalidTemplates{}, &authRepo{}, storage, cellphoneInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUpdateCellHandler(t *testing.T) {
	storage := new(cellphoneStorage)
	storage.On("UpdateCell", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler := profiles.CreateUpdateCellHandler(&templates{}, &authRepo{}, storage, cellphoneInputFn)
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
