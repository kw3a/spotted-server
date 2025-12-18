package quizestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type keystrokeReportStorageMock struct {
	mock.Mock
}

func (m *keystrokeReportStorageMock) SelectStrokeWindows(ctx context.Context, participationID string) ([]shared.StrokeWindow, error) {
	args := m.Called(ctx, participationID)
	return args.Get(0).([]shared.StrokeWindow), args.Error(1)
}

func TestKeyStrokeReportHandlerBadInput(t *testing.T) {
	badInputFn := func(r *http.Request) (quizes.KeyStrokeReportInput, error) {
		return quizes.KeyStrokeReportInput{}, errors.New("bad input")
	}

	handler := quizes.CreateKeyStrokeReportHandler(
		&templates{},
		&keystrokeReportStorageMock{},
		badInputFn,
	)

	req, _ := http.NewRequest("GET", "/report", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKeyStrokeReportHandlerStorageError(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.KeyStrokeReportInput, error) {
		return quizes.KeyStrokeReportInput{ParticipationID: "part-id"}, nil
	}

	storage := new(keystrokeReportStorageMock)
	storage.On("SelectStrokeWindows", mock.Anything, "part-id").Return([]shared.StrokeWindow{}, errors.New("storage error"))

	handler := quizes.CreateKeyStrokeReportHandler(
		&templates{},
		storage,
		inputFn,
	)

	req, _ := http.NewRequest("GET", "/report", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestKeyStrokeReportHandlerRenderError(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.KeyStrokeReportInput, error) {
		return quizes.KeyStrokeReportInput{ParticipationID: "part-id"}, nil
	}

	storage := new(keystrokeReportStorageMock)
	storage.On("SelectStrokeWindows", mock.Anything, "part-id").Return([]shared.StrokeWindow{}, nil)

	handler := quizes.CreateKeyStrokeReportHandler(
		&invalidTemplates{},
		storage,
		inputFn,
	)

	req, _ := http.NewRequest("GET", "/report", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestKeyStrokeReportHandlerSuccess(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.KeyStrokeReportInput, error) {
		return quizes.KeyStrokeReportInput{ParticipationID: "part-id"}, nil
	}

	storage := new(keystrokeReportStorageMock)
	windows := []shared.StrokeWindow{
		{StrokeAmount: 100, UdMean: 10},
		{StrokeAmount: 50, UdMean: 12},
	}
	storage.On("SelectStrokeWindows", mock.Anything, "part-id").Return(windows, nil)

	handler := quizes.CreateKeyStrokeReportHandler(
		&templates{},
		storage,
		inputFn,
	)

	req, _ := http.NewRequest("GET", "/report", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
