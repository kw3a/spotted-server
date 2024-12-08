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

type ExperienceRegisterInput struct {
	Company string
	Title   string
	Start   time.Time
	End     time.Time
}

type ExperienceDeleteInput struct {
	ExperienceID string
}


func GetExperienceRegisterInput(r *http.Request) (ExperienceRegisterInput, error) {
	company := r.FormValue("company")
	if company == "" {
		return ExperienceRegisterInput{}, fmt.Errorf("company is required")
	}
	title := r.FormValue("title")
	if title == "" {
		return ExperienceRegisterInput{}, fmt.Errorf("title is required")
	}
	start, err := time.Parse("2006-01", r.FormValue("start"))
	if err != nil {
		return ExperienceRegisterInput{}, fmt.Errorf("start is required")
	}
	end, err := time.Parse("2006-01", r.FormValue("end"))
	if err != nil {
		return ExperienceRegisterInput{}, fmt.Errorf("end is required")
	}
	return ExperienceRegisterInput{
		Company: company,
		Title:   title,
		Start:   start,
		End:     end,
	}, nil
}

func GetExperienceDeleteInput(r *http.Request) (ExperienceDeleteInput, error) {
	experienceID := chi.URLParam(r, "experienceID")
	if err := shared.ValidateUUID(experienceID); err != nil {
		return ExperienceDeleteInput{}, err
	}
	return ExperienceDeleteInput{
		ExperienceID: experienceID,
	}, nil
}

type ExperienceStorage interface {
	RegisterExperience(ctx context.Context, experienceID, userID, company, title string, start, end time.Time) error
	DeleteExperience(ctx context.Context, userID, experienceID string) error
}

type registerExperienceInputFn func(r *http.Request) (ExperienceRegisterInput, error)

func CreateRegisterExperienceHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage ExperienceStorage, inputFn registerExperienceInputFn) http.HandlerFunc {
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
		experienceID := uuid.New().String()
		if err := storage.RegisterExperience(r.Context(), experienceID, user.ID, input.Company, input.Title, input.Start, input.End); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := shared.ExperienceEntry{
			Company:     input.Company,
			Title:       input.Title,
			StartDate:   input.Start.Format("2006-01"),
			EndDate:     input.End.Format("2006-01"),
			ID:          experienceID,
		}
		if err := templ.Render(w, "experienceEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type deleteExperienceInputFn func(r *http.Request) (ExperienceDeleteInput, error)

func CreateDeleteExperienceHandler(auth shared.AuthRep, storage ExperienceStorage, inputFn deleteExperienceInputFn) http.HandlerFunc {
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
		if err := storage.DeleteExperience(r.Context(), user.ID, input.ExperienceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

