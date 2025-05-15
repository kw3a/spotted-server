package profiles

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type SkillRegisterInput struct {
	Name string
}

type SkillRegisterErrors struct {
	NameError string
}

type SkillDeleteInput struct {
	SkillID string
}


func GetSkillRegisterInput(r *http.Request) (SkillRegisterInput, SkillRegisterErrors, bool) {
	name := r.FormValue("name")
	errFound := false
	inputErrors := SkillRegisterErrors{}
	if name == "" {
		errFound = true
		inputErrors.NameError = errNameRequired
	}
	return SkillRegisterInput{
		Name: name,
	}, inputErrors, errFound
}

func GetSkillDeleteInput(r *http.Request) (SkillDeleteInput, error) {
	skillID := chi.URLParam(r, "skillID")
	if err := shared.ValidateUUID(skillID); err != nil {
		return SkillDeleteInput{}, err
	}
	return SkillDeleteInput{
		SkillID: skillID,
	}, nil
}

type SkillStorage interface {
	RegisterSkill(ctx context.Context, skillID, userID, name string) error
	DeleteSkill(ctx context.Context, userID, skillID string) error
}

type registerSkillInputFn func(r *http.Request) (SkillRegisterInput, SkillRegisterErrors, bool)

func CreateRegisterSkillHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage SkillStorage, inputFn registerSkillInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErrors, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "skillErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		skillID := uuid.New().String()
		if err := storage.RegisterSkill(r.Context(), skillID, user.ID, input.Name); err != nil {
			inputErrors.NameError = errUnexpected
			if err := templ.Render(w, "skillErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		data := shared.SkillEntry{
			Name: input.Name,
			ID:   skillID,
		}
		w.Header().Set("HX-Trigger", "skill-added")
		if err := templ.Render(w, "skillEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type deleteSkillInputFn func(r *http.Request) (SkillDeleteInput, error)

func CreateDeleteSkillHandler(auth shared.AuthRep, storage SkillStorage, inputFn deleteSkillInputFn) http.HandlerFunc {
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
		if err := storage.DeleteSkill(r.Context(), user.ID, input.SkillID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

