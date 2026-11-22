package quizes

import (
	"context"
	"math"
	"net/http"
	"sort"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type KeyStrokeReportStorage interface {
	SelectStrokeWindows(ctx context.Context, participationID string) ([]shared.StrokeWindow, error)
}

type KeyStrokeReportInput struct{ ParticipationID string }

func GetKeyStrokeReportInput(r *http.Request) (KeyStrokeReportInput, error) {
	participationID := r.FormValue("participationID")
	if err := shared.ValidateUUID(participationID); err != nil {
		return KeyStrokeReportInput{}, err
	}
	return KeyStrokeReportInput{ParticipationID: participationID}, nil
}

type keystrokeReportInputFn func(*http.Request) (KeyStrokeReportInput, error)

func CreateKeyStrokeReportHandler(templ shared.TemplatesRepo, storage KeyStrokeReportStorage, inputFn keystrokeReportInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		windows, err := storage.SelectStrokeWindows(r.Context(), input.ParticipationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := templ.Render(w, "keystrokeReport", AnalyzeKeystrokes(windows, input.ParticipationID)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type AnalysisPoint struct {
	WindowIndex    int     `json:"windowIndex"`
	Score          float64 `json:"score"`
	Cusum          float64 `json:"cusum"`
	IsBaseline     bool    `json:"isBaseline"`
	IsSuspicious   bool    `json:"isSuspicious"`
	IsInsufficient bool    `json:"isInsufficient"`
	HasGapBefore   bool    `json:"hasGapBefore"`
	StrokeCount    int32   `json:"strokeCount"`
	StartAt        int64   `json:"startAt"`
	EndAt          int64   `json:"endAt"`
}

type KeystrokeReportData struct {
	Points          []AnalysisPoint `json:"points"`
	ParticipationID string
}

const (
	strokeWindowInterval = 3 * 60
	minimumRobustStrokes = 50
	baselineWindows      = 3
	cusumAllowance       = 1.5
	cusumThreshold       = 4.0
)

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func robustFeatures(w shared.StrokeWindow) []float64 {
	return []float64{float64(w.UdMedian), float64(w.Du1Median), float64(w.Du2Median), float64(w.DdMedian), float64(w.UuMedian), float64(w.UdMad), float64(w.Du1Mad), float64(w.Du2Mad), float64(w.DdMad), float64(w.UuMad), float64(w.LongPausePermille)}
}

func windowTimes(w shared.StrokeWindow) (int64, int64) {
	if w.FeatureVersion > 0 && w.WindowStartMs > 0 && w.WindowEndMs >= w.WindowStartMs {
		return int64(w.WindowStartMs / 1000), int64(w.WindowEndMs / 1000)
	}
	end := w.CreatedAt.Unix()
	return end - strokeWindowInterval, end
}

func windowOrderTimes(w shared.StrokeWindow) (int64, int64) {
	if w.FeatureVersion > 0 && w.WindowStartMs > 0 && w.WindowEndMs >= w.WindowStartMs {
		return int64(w.WindowStartMs), int64(w.WindowEndMs)
	}
	end := w.CreatedAt.UnixMilli()
	return end - int64(strokeWindowInterval)*1000, end
}

func AnalyzeKeystrokes(windows []shared.StrokeWindow, participationID string) KeystrokeReportData {
	sort.SliceStable(windows, func(i, j int) bool {
		startI, endI := windowOrderTimes(windows[i])
		startJ, endJ := windowOrderTimes(windows[j])
		if startI != startJ {
			return startI < startJ
		}
		if endI != endJ {
			return endI < endJ
		}
		return windows[i].ID < windows[j].ID
	})
	points := make([]AnalysisPoint, 0, len(windows))
	validIndexes := make([]int, 0)
	for i, w := range windows {
		start, end := windowTimes(w)
		p := AnalysisPoint{WindowIndex: i, StrokeCount: w.StrokeAmount, StartAt: start, EndAt: end}
		p.IsInsufficient = w.FeatureVersion != 1 || w.StrokeAmount < minimumRobustStrokes || end < start
		if i > 0 {
			_, previousEnd := windowTimes(windows[i-1])
			p.HasGapBefore = start-previousEnd > strokeWindowInterval+30
		}
		if !p.IsInsufficient {
			validIndexes = append(validIndexes, i)
		}
		points = append(points, p)
	}
	if len(validIndexes) < baselineWindows {
		return KeystrokeReportData{Points: points, ParticipationID: participationID}
	}

	center := make([]float64, 11)
	for feature := range center {
		values := make([]float64, baselineWindows)
		for j := 0; j < baselineWindows; j++ {
			values[j] = robustFeatures(windows[validIndexes[j]])[feature]
		}
		center[feature] = medianFloat(values)
	}
	scale := make([]float64, len(center))
	for feature := range scale {
		deviations := make([]float64, baselineWindows)
		measurementFloor := 1.0
		for j := 0; j < baselineWindows; j++ {
			w := windows[validIndexes[j]]
			features := robustFeatures(w)
			deviations[j] = math.Abs(features[feature] - center[feature])
			if feature < 5 {
				withinMAD := robustFeatures(w)[feature+5]
				measurementFloor = math.Max(measurementFloor, 1.4826*1.253*withinMAD/math.Sqrt(float64(w.StrokeAmount)))
			} else if feature < 10 {
				withinMAD := robustFeatures(w)[feature]
				measurementFloor = math.Max(measurementFloor, 2.0*withinMAD/math.Sqrt(float64(w.StrokeAmount)))
			} else {
				p := float64(w.LongPausePermille) / 1000
				measurementFloor = math.Max(measurementFloor, 1000*math.Sqrt(p*(1-p)/float64(w.StrokeAmount)))
			}
		}
		scale[feature] = math.Max(measurementFloor, 1.4826*medianFloat(deviations))
	}
	baseline := make(map[int]bool, baselineWindows)
	for _, index := range validIndexes[:baselineWindows] {
		baseline[index] = true
		points[index].IsBaseline = true
	}
	consecutive := 0.0
	var previousEnd int64
	excursionStart := -1
	for i, w := range windows {
		point := &points[i]
		if point.IsInsufficient {
			consecutive = 0
			excursionStart = -1
			continue
		}
		if point.HasGapBefore || (previousEnd > 0 && point.StartAt-previousEnd > strokeWindowInterval+30) {
			consecutive = 0
			excursionStart = -1
		}
		previousEnd = point.EndAt
		if baseline[i] {
			point.Cusum = 0
			excursionStart = -1
			continue
		}
		features := robustFeatures(w)
		squared := 0.0
		for j, value := range features {
			z := math.Min(3, math.Abs(value-center[j])/scale[j])
			weight := 1.0
			if j == 10 {
				weight = 0.5
			}
			squared += weight * z * z
		}
		point.Score = math.Sqrt(squared / 10.5)
		consecutive = math.Max(0, consecutive+point.Score-cusumAllowance)
		point.Cusum = consecutive
		if consecutive == 0 {
			excursionStart = -1
		} else if excursionStart == -1 {
			excursionStart = i
		}
		if consecutive >= cusumThreshold {
			for j := excursionStart; j <= i; j++ {
				points[j].IsSuspicious = true
			}
		}
	}
	return KeystrokeReportData{Points: points, ParticipationID: participationID}
}
