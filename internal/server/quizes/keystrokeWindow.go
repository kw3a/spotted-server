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
	parseOptional := func(key string) (int32, error) {
		if r.FormValue(key) == "" {
			return 0, nil
		}
		return parse(key)
	}

	var input StrokeWindowInput
	var sw shared.StrokeWindow
	quizID := r.FormValue("quizID")
	if err := shared.ValidateUUID(quizID); err != nil {
		return StrokeWindowInput{}, err
	}
	input.QuizID = quizID
	windowID := r.FormValue("windowId")
	if windowID != "" {
		if err := shared.ValidateUUID(windowID); err != nil {
			return StrokeWindowInput{}, err
		}
		sw.ID = windowID
	}
	featureVersion := int32(0)
	var err error
	if r.FormValue("featureVersion") != "" {
		featureVersion, err = parse("featureVersion")
		if err != nil {
			return StrokeWindowInput{}, err
		}
	}
	if featureVersion < 0 || featureVersion > 1 {
		return StrokeWindowInput{}, strconv.ErrRange
	}
	sw.FeatureVersion = uint8(featureVersion)
	parseUint64 := func(key string) (uint64, error) {
		value, err := strconv.ParseUint(r.FormValue(key), 10, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	if r.FormValue("windowStartMs") != "" {
		if sw.WindowStartMs, err = parseUint64("windowStartMs"); err != nil {
			return StrokeWindowInput{}, err
		}
	}
	if r.FormValue("windowEndMs") != "" {
		if sw.WindowEndMs, err = parseUint64("windowEndMs"); err != nil {
			return StrokeWindowInput{}, err
		}
	}
	if sw.FeatureVersion > 0 && (sw.WindowStartMs == 0 || sw.WindowEndMs < sw.WindowStartMs) {
		return StrokeWindowInput{}, strconv.ErrRange
	}
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
	if sw.UdMedian, err = parseOptional("udMedian"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.UdMad, err = parseOptional("udMad"); err != nil {
		return StrokeWindowInput{}, err
	}
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
	if sw.Du1Median, err = parseOptional("du1Median"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.Du1Mad, err = parseOptional("du1Mad"); err != nil {
		return StrokeWindowInput{}, err
	}
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
	if sw.Du2Median, err = parseOptional("du2Median"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.Du2Mad, err = parseOptional("du2Mad"); err != nil {
		return StrokeWindowInput{}, err
	}
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
	if sw.DdMedian, err = parseOptional("ddMedian"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.DdMad, err = parseOptional("ddMad"); err != nil {
		return StrokeWindowInput{}, err
	}
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
	if sw.UuMedian, err = parseOptional("uuMedian"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.UuMad, err = parseOptional("uuMad"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.LongPausePermille, err = parseOptional("longPausePermille"); err != nil {
		return StrokeWindowInput{}, err
	}
	if sw.LongPausePermille < 0 || sw.LongPausePermille > 1000 {
		return StrokeWindowInput{}, strconv.ErrRange
	}
	input.StrokeWindow = sw
	return input, nil
}

type keystrokeWindowInputFn func(*http.Request) (StrokeWindowInput, error)

const telemetryWindowTolerance = 2 * time.Minute

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
		if input.StrokeWindow.FeatureVersion > 0 {
			const maxInt64 = int64(1<<63 - 1)
			if input.StrokeWindow.WindowStartMs > uint64(maxInt64) || input.StrokeWindow.WindowEndMs > uint64(maxInt64) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			oldestWindow := partiData.CreatedAt.Add(-telemetryWindowTolerance).UnixMilli()
			latestWindow := partiData.ExpiresAt.Add(telemetryWindowTolerance).UnixMilli()
			if int64(input.StrokeWindow.WindowStartMs) < oldestWindow || int64(input.StrokeWindow.WindowEndMs) > latestWindow {
				// Return success so the client discards stale queued telemetry instead of retrying it forever.
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		if input.StrokeWindow.ID == "" {
			input.StrokeWindow.ID = uuid.NewString()
		}
		if err := storage.InsertKeyStrokeWindow(r.Context(), partiData.ID, input.StrokeWindow); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
