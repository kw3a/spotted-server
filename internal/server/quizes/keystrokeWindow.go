package quizes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type KeyStrokeWindowStorage interface {
	InsertKeyStrokeWindow(
		ctx context.Context,
		participationID string,
		strokeWindow shared.StrokeWindow,
	) error
	ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error)
}

type StrokeWindowInput struct {
	StrokeWindow shared.StrokeWindow
	QuizID       string
}

func GetStrokeWindowInput(r *http.Request) (StrokeWindowInput, error) {
	parse := func(key string) (int32, error) {
		valStr := r.FormValue(key)
		val64, err := strconv.Atoi(valStr)
		if err != nil {
			return 0, err
		}
		return shared.IntToInt32(val64), nil
	}

	var input StrokeWindowInput
	var sw shared.StrokeWindow
	quizID := r.FormValue("quizID")
	if err := shared.ValidateUUID(quizID); err != nil {
		return StrokeWindowInput{}, err
	}
	input.QuizID = quizID
	strokeAmount, err := parse("strokeAmount")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.StrokeAmount = strokeAmount
	udMean, err := parse("udMean")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.UdMean = udMean
	udStdDev, err := parse("udStdDev")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.UdStdDev = udStdDev
	du1Mean, err := parse("du1Mean")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.Du1Mean = du1Mean
	du1StdDev, err := parse("du1StdDev")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.Du1StdDev = du1StdDev
	du2Mean, err := parse("du2Mean")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.Du2Mean = du2Mean
	du2StdDev, err := parse("du2StdDev")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.Du2StdDev = du2StdDev
	ddMean, err := parse("ddMean")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.DdMean = ddMean
	ddStdDev, err := parse("ddStdDev")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.DdStdDev = ddStdDev
	uuMean, err := parse("uuMean")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.UuMean = uuMean
	uuStdDev, err := parse("uuStdDev")
	if err != nil {
		return StrokeWindowInput{}, err
	}
	sw.UuStdDev = uuStdDev
	input.StrokeWindow = sw
	return input, nil
}

type keystrokeWindowInputFn func(*http.Request) (StrokeWindowInput, error)

func CreateKeyStrokeWindowHandler(storage KeyStrokeWindowStorage, authService shared.AuthRep, inputFn keystrokeWindowInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		partiData, err := storage.ParticipationStatus(r.Context(), user.ID, input.QuizID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if partiData.ExpiresAt.Before(time.Now()) {
			http.Error(w, "your participation is over", http.StatusUnauthorized)
			return
		}
		input.StrokeWindow.ID = uuid.NewString()
		if err := storage.InsertKeyStrokeWindow(r.Context(), partiData.ID, input.StrokeWindow); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
