package quizes

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errNoSrc = "No se ha recibido ninguna solución"
)

type SrcData struct {
	Submission  shared.Submission
	TestCases   []shared.ExecutedTestCase
	ApplicantID string
	ProblemID   string
}

type NoSrcData struct {
	Msg         string
	ApplicantID string
	ProblemID   string
}

type SrcInput struct {
	ApplicantID string
	ProblemID   string
}

type SrcStorage interface {
	BestSubmission(ctx context.Context, applicantID, problemID string) (shared.Submission, error)
	GetExecutedTestCases(ctx context.Context, problemID, userID string) ([]shared.ExecutedTestCase, error)
}

func GetSrcInput(r *http.Request) (SrcInput, error) {
	applicantID := chi.URLParam(r, "applicantID")
	if err := shared.ValidateUUID(applicantID); err != nil {
		return SrcInput{}, err
	}
	problemID := chi.URLParam(r, "problemID")
	if err := shared.ValidateUUID(problemID); err != nil {
		return SrcInput{}, err
	}
	return SrcInput{
		ApplicantID: applicantID,
		ProblemID:   problemID,
	}, nil
}

type srcInputFn func(r *http.Request) (SrcInput, error)

func CreateSrcHandler(
	inputFn srcInputFn,
	storage SrcStorage,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		subm, err := storage.BestSubmission(r.Context(), input.ApplicantID, input.ProblemID)
		if err != nil {
			if err == sql.ErrNoRows {
				if err := templ.Render(w, "noSourceCard", NoSrcData{
					Msg:         errNoSrc,
					ApplicantID: input.ApplicantID,
					ProblemID:   input.ProblemID,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		executedTCs, err := storage.GetExecutedTestCases(r.Context(), input.ProblemID, subm.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := templ.Render(w, "sourceCard", SrcData{
			Submission:  subm,
			TestCases:   executedTCs,
			ApplicantID: input.ApplicantID,
			ProblemID:   input.ProblemID,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
