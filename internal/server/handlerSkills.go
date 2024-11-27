package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SkillRegisterInput struct {
	Name string
}

type SkillDeleteInput struct {
	SkillID string
}

type SkillEntry struct {
	Name string
	ID   string
}

func GetSkillRegisterInput(r *http.Request) (SkillRegisterInput, error) {
	name := r.FormValue("name")
	if name == "" {
		return SkillRegisterInput{}, fmt.Errorf("name is required")
	}
	return SkillRegisterInput{
		Name: name,
	}, nil
}

func GetSkillDeleteInput(r *http.Request) (SkillDeleteInput, error) {
	skillID := chi.URLParam(r, "skillID")
	if err := ValidateUUID(skillID); err != nil {
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

type registerSkillInputFn func(r *http.Request) (SkillRegisterInput, error)

func CreateRegisterSkillHandler(templ TemplatesRepo, auth AuthRep, storage SkillStorage, inputFn registerSkillInputFn) http.HandlerFunc {
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
		skillID := uuid.New().String()
		if err := storage.RegisterSkill(r.Context(), skillID, user.ID, input.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := SkillEntry{
			Name: input.Name,
			ID:   skillID,
		}
		if err := templ.Render(w, "skillEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type deleteSkillInputFn func(r *http.Request) (SkillDeleteInput, error)

func CreateDeleteSkillHandler(auth AuthRep, storage SkillStorage, inputFn deleteSkillInputFn) http.HandlerFunc {
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

func (DI *App) SkillRegisterHandler() http.HandlerFunc {
	return CreateRegisterSkillHandler(DI.Templ, DI.AuthService, DI.Storage, GetSkillRegisterInput)
}

func (DI *App) SkillDeleteHandler() http.HandlerFunc {
	return CreateDeleteSkillHandler(DI.AuthService, DI.Storage, GetSkillDeleteInput)
}
