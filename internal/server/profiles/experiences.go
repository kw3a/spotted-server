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

const (
	errInvalidDate = "Fecha inválida"
)

type ExpRegInput struct {
	Company string
	Title   string
	Start   time.Time
	End     sql.NullTime
}

type ExperienceDeleteInput struct {
	ExperienceID string
}

type ExpRegErrors struct {
	TitleError string
	CompError  string
	StartError string
	EndError   string
}

func GetExperienceRegisterInput(r *http.Request) (ExpRegInput, ExpRegErrors, bool) {
	inputErrors := ExpRegErrors{}
	errFound := false
	company := r.FormValue("company")
	if len(company) > 64 || company == "" {
		inputErrors.CompError = shared.ErrLength(1, 64)
		errFound = true
	}
	title := r.FormValue("title")
	if len(title) < 5 || len(title) > 256 {
		inputErrors.TitleError = shared.ErrLength(5, 256)
		errFound = true
	}
	start, err := time.Parse("2006-01", r.FormValue("start"))
	if err != nil {
		inputErrors.StartError = errInvalidDate
		errFound = true
	}
	dateEnd := sql.NullTime{Valid: false}
	end := r.FormValue("end")
	if end != "" {
		parsedEnd, err := time.Parse("2006-01", end)
		if err != nil {
			inputErrors.EndError = errInvalidDate
			errFound = true
		} else {
			dateEnd.Valid = true
			dateEnd.Time = parsedEnd
		}
	}
	return ExpRegInput{
		Company: company,
		Title:   title,
		Start:   start,
		End:     dateEnd,
	}, inputErrors, errFound
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
	RegisterExperience(
		ctx context.Context,
		experienceID, userID, company, title string,
		start time.Time,
		end sql.NullTime,
	) error
	DeleteExperience(ctx context.Context, userID, experienceID string) error
}

type registerExperienceInputFn func(r *http.Request) (ExpRegInput, ExpRegErrors, bool)

func CreateRegisterExperienceHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage ExperienceStorage, inputFn registerExperienceInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			renderErr := templ.Render(w, "expErrors", inputErr)
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		experienceID := uuid.New().String()
		if err := storage.RegisterExperience(
			r.Context(),
			experienceID,
			user.ID,
			input.Company,
			input.Title,
			input.Start,
			input.End,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := shared.ExperienceEntry{
			Company:      input.Company,
			Title:        input.Title,
			StartDate:    shared.DateSpanishFormat(sql.NullTime{Valid: true, Time: input.Start}),
			EndDate:      shared.DateSpanishFormat(input.End),
			TimeInterval: shared.TimeInterval(input.Start, input.End),
			ID:           experienceID,
		}
		w.Header().Set("HX-Trigger", "exp-added")
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
