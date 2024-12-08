package profiles

import (
	"context"
	"fmt"
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
	End         time.Time
}

type EducationDeleteInput struct {
	EducationID string
}

func GetEducationRegisterInput(r *http.Request) (EducationRegisterInput, error) {
	institution := r.FormValue("institution")
	if institution == "" {
		return EducationRegisterInput{}, fmt.Errorf("institution is required")
	}
	degree := r.FormValue("degree")
	if degree == "" {
		return EducationRegisterInput{}, fmt.Errorf("degree is required")
	}
	start, err := time.Parse("2006-01", r.FormValue("start"))
	if err != nil {
		return EducationRegisterInput{}, fmt.Errorf("start is required")
	}
	end, err := time.Parse("2006-01", r.FormValue("end"))
	if err != nil {
		return EducationRegisterInput{}, fmt.Errorf("end is required")
	}
	return EducationRegisterInput{
		Institution: institution,
		Degree:      degree,
		Start:       start,
		End:         end,
	}, nil
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
	RegisterEducation(ctx context.Context, educationID, userID, institution, degree string, start, end time.Time) error
	DeleteEducation(ctx context.Context, userID, educationID string) error
}

type registerEducationInputFn func(r *http.Request) (EducationRegisterInput, error)

func CreateRegisterEducationHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage EducationStorage, inputFn registerEducationInputFn) http.HandlerFunc {
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
		educationID := uuid.New().String()
		if err := storage.RegisterEducation(r.Context(), educationID, user.ID, input.Institution, input.Degree, input.Start, input.End); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := shared.EducationEntry{
			Institution: input.Institution,
			Degree:      input.Degree,
			StartDate:   input.Start.Format("2006-01"),
			EndDate:     input.End.Format("2006-01"),
			ID:          educationID,
		}
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
		if err := storage.DeleteEducation(r.Context(), user.ID, input.EducationID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

