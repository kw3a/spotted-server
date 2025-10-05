package profiles

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type EducationRegisterInput struct {
	Institution string
	Degree      string
	Start       time.Time
	End         sql.NullTime
}

type EducationDeleteInput struct {
	EducationID string
}

type EducationRegErrors struct {
	InstitutionError string
	DegreeError      string
	StartError       string
	EndError         string
}

func GetEducationRegisterInput(r *http.Request) (EducationRegisterInput, EducationRegErrors, bool) {
	inputErrors := EducationRegErrors{}
	errFound := false

	institution := r.FormValue("institution")
	if len(institution) < 5 || len(institution) > 128 {
		inputErrors.InstitutionError = shared.ErrLength(5, 128)
		errFound = true
	}

	degree := r.FormValue("degree")
	if len(degree) < 5 || len(degree) > 128 {
		inputErrors.DegreeError = shared.ErrLength(5, 128)
		errFound = true
	}

	start, err := time.Parse("2006-01", r.FormValue("start"))
	if err != nil {
		inputErrors.StartError = errInvalidDate
		errFound = true
	}

	end := r.FormValue("end")
	dateEnd := sql.NullTime{Valid: false}
	if end != "" {
		parsedEnd, err := time.Parse("2006-01", end)
		if err != nil {
			inputErrors.EndError = errInvalidDate
			errFound = true
		} else {
			dateEnd = sql.NullTime{Time: parsedEnd, Valid: true}
		}
	}

	return EducationRegisterInput{
		Institution: institution,
		Degree:      degree,
		Start:       start,
		End:         dateEnd,
	}, inputErrors, errFound
}

func GetEducationDeleteInput(r *http.Request) (EducationDeleteInput, error) {
	educationID := chi.URLParam(r, "educationID")
	if err := shared.ValidateUUID(educationID); err != nil {
		return EducationDeleteInput{}, err
	}
	return EducationDeleteInput{
		EducationID: educationID,
	}, nil
}

type EducationStorage interface {
	CountEducation(ctx context.Context, userID string) (int32, error)
	RegisterEducation(ctx context.Context, educationID, userID, institution, degree string, start time.Time, end sql.NullTime) error
	DeleteEducation(ctx context.Context, userID, educationID string) error
}

type registerEducationInputFn func(r *http.Request) (EducationRegisterInput, EducationRegErrors, bool)

func CreateRegisterEducationHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage EducationStorage, inputFn registerEducationInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "edErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		educationID := uuid.New().String()
		count, err := storage.CountEducation(r.Context(), user.ID)
		if err != nil || count >= maxGroupSize {
			inputErr.DegreeError = errGroupSize(maxGroupSize)
		} else {
			if err := storage.RegisterEducation(r.Context(), educationID, user.ID, input.Institution, input.Degree, input.Start, input.End); err != nil {
				inputErr.DegreeError = errUnexpected
			}
		}
		if inputErr.DegreeError != "" {
			if err := templ.Render(w, "edErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		data := shared.EducationEntry{
			Institution: input.Institution,
			Degree:      input.Degree,
			StartDate:   shared.DateSpanishFormat(sql.NullTime{Time: input.Start, Valid: true}),
			EndDate:     shared.DateSpanishFormat(input.End),
			ID:          educationID,
		}
		w.Header().Set("HX-Trigger", "ed-added")
		if err := templ.Render(w, "educationEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type deleteEducationInputFn func(r *http.Request) (EducationDeleteInput, error)

func CreateDeleteEducationHandler(auth shared.AuthRep, storage EducationStorage, inputFn deleteEducationInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = storage.DeleteEducation(r.Context(), user.ID, input.EducationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
