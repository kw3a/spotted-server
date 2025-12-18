package quizestest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type keystrokeWindowStorageMock struct {
	mock.Mock
}

func (m *keystrokeWindowStorageMock) InsertKeyStrokeWindow(ctx context.Context, participationID string, strokeWindow shared.StrokeWindow) error {
	args := m.Called(ctx, participationID, strokeWindow)
	return args.Error(0)
}

func (m *keystrokeWindowStorageMock) ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error) {
	args := m.Called(ctx, userID, quizID)
	return args.Get(0).(shared.Participation), args.Error(1)
}

func TestKeyStrokeWindowHandlerUnauthorized(t *testing.T) {
	handler := quizes.CreateKeyStrokeWindowHandler(
		&keystrokeWindowStorageMock{},
		&invalidAuthRepo{},
		nil,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestKeyStrokeWindowHandlerBadInput(t *testing.T) {
	badInputFn := func(r *http.Request) (quizes.StrokeWindowInput, error) {
		return quizes.StrokeWindowInput{}, errors.New("bad input")
	}

	handler := quizes.CreateKeyStrokeWindowHandler(
		&keystrokeWindowStorageMock{},
		&authRepo{},
		badInputFn,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKeyStrokeWindowHandlerParticipationError(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.StrokeWindowInput, error) {
		return quizes.StrokeWindowInput{QuizID: "quiz-id"}, nil
	}

	storage := new(keystrokeWindowStorageMock)
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, "quiz-id").Return(shared.Participation{}, errors.New("participation error"))

	handler := quizes.CreateKeyStrokeWindowHandler(
		storage,
		&authRepo{},
		inputFn,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKeyStrokeWindowHandlerExpired(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.StrokeWindowInput, error) {
		return quizes.StrokeWindowInput{QuizID: "quiz-id"}, nil
	}

	storage := new(keystrokeWindowStorageMock)
	expiredParticipation := shared.Participation{
		ID:        "part-id",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, "quiz-id").Return(expiredParticipation, nil)

	handler := quizes.CreateKeyStrokeWindowHandler(
		storage,
		&authRepo{},
		inputFn,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "your participation is over")
}

func TestKeyStrokeWindowHandlerInsertError(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.StrokeWindowInput, error) {
		return quizes.StrokeWindowInput{QuizID: "quiz-id", StrokeWindow: shared.StrokeWindow{StrokeAmount: 10}}, nil
	}

	storage := new(keystrokeWindowStorageMock)
	activeParticipation := shared.Participation{
		ID:        "part-id",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, "quiz-id").Return(activeParticipation, nil)
	storage.On("InsertKeyStrokeWindow", mock.Anything, "part-id", mock.MatchedBy(func(sw shared.StrokeWindow) bool {
		return sw.StrokeAmount == 10
	})).Return(errors.New("insert error"))

	handler := quizes.CreateKeyStrokeWindowHandler(
		storage,
		&authRepo{},
		inputFn,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestKeyStrokeWindowHandlerSuccess(t *testing.T) {
	inputFn := func(r *http.Request) (quizes.StrokeWindowInput, error) {
		return quizes.StrokeWindowInput{QuizID: "quiz-id", StrokeWindow: shared.StrokeWindow{StrokeAmount: 10}}, nil
	}

	storage := new(keystrokeWindowStorageMock)
	activeParticipation := shared.Participation{
		ID:        "part-id",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	storage.On("ParticipationStatus", mock.Anything, mock.Anything, "quiz-id").Return(activeParticipation, nil)
	storage.On("InsertKeyStrokeWindow", mock.Anything, "part-id", mock.MatchedBy(func(sw shared.StrokeWindow) bool {
		return sw.StrokeAmount == 10
	})).Return(nil)

	handler := quizes.CreateKeyStrokeWindowHandler(
		storage,
		&authRepo{},
		inputFn,
	)

	req, _ := http.NewRequest("POST", "/window", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
