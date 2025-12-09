package quizes

import (
	"context"
	"math"
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type KeyStrokeReportStorage interface {
	SelectStrokeWindows(ctx context.Context, participationID string) ([]shared.StrokeWindow, error)
}

type KeyStrokeReportInput struct {
	ParticipationID string
}

func GetKeyStrokeReportInput(r *http.Request) (KeyStrokeReportInput, error) {
	participationID := r.FormValue("participationID")
	err := shared.ValidateUUID(participationID)
	if err != nil {
		return KeyStrokeReportInput{}, err
	}
	return KeyStrokeReportInput{ParticipationID: participationID}, nil
}

type keystrokeReportInputFn func(*http.Request) (KeyStrokeReportInput, error)

func CreateKeyStrokeReportHandler(
	templ shared.TemplatesRepo,
	storage KeyStrokeReportStorage,
	inputFn keystrokeReportInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ksWindows, err := storage.SelectStrokeWindows(r.Context(), input.ParticipationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reportData := AnalyzeKeystrokes(ksWindows, input.ParticipationID)
		if err := templ.Render(w, "keystrokeReport", reportData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type AnalysisPoint struct {
	WindowIndex int     `json:"windowIndex"`
	SMD         float64 `json:"smd"`
	IsProfile   bool    `json:"isProfile"`
	IsInactive  bool    `json:"isInactive"`
	StrokeCount int32   `json:"strokeCount"`
}

type KeystrokeReportData struct {
	Points          []AnalysisPoint `json:"points"`
	ParticipationID string
}

func AnalyzeKeystrokes(windows []shared.StrokeWindow, participationID string) KeystrokeReportData {
	if len(windows) < 2 {
		return KeystrokeReportData{Points: []AnalysisPoint{}, ParticipationID: participationID}
	}

	// 1. Find profile window (max strokes)
	profileIndex := 0
	maxStrokes := int32(0)
	for i, w := range windows {
		if w.StrokeAmount > maxStrokes {
			maxStrokes = w.StrokeAmount
			profileIndex = i
		}
	}

	profile := windows[profileIndex]
	var points []AnalysisPoint

	// 2. Calculate SMD for each window
	for i, w := range windows {
		point := AnalysisPoint{
			WindowIndex: i,
			StrokeCount: w.StrokeAmount,
		}

		if i == profileIndex {
			point.IsProfile = true
			points = append(points, point)
			continue
		}

		if w.StrokeAmount < 10 {
			point.IsInactive = true
			points = append(points, point)
			continue
		}

		// Calculate SMD
		// Features: Ud, Du1, Du2, Dd, Uu
		// SMD = Sum(|Mean_p - Mean_t| / StdDev_p) / N
		// Skip if StdDev_p is 0

		sumDist := 0.0
		validFeatures := 0

		// Helper for one feature
		calcFeature := func(pMean, pStd, tMean int32) {
			if pStd == 0 {
				return
			}
			dist := math.Abs(float64(pMean-tMean)) / float64(pStd)
			sumDist += dist
			validFeatures++
		}

		calcFeature(profile.UdMean, profile.UdStdDev, w.UdMean)
		calcFeature(profile.Du1Mean, profile.Du1StdDev, w.Du1Mean)
		calcFeature(profile.Du2Mean, profile.Du2StdDev, w.Du2Mean)
		calcFeature(profile.DdMean, profile.DdStdDev, w.DdMean)
		calcFeature(profile.UuMean, profile.UuStdDev, w.UuMean)

		if validFeatures > 0 {
			point.SMD = sumDist / float64(validFeatures)
		}

		points = append(points, point)
	}

	return KeystrokeReportData{Points: points, ParticipationID: participationID}
}
