package profiles

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/auth"
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

const maxGroupSize = 10

func errGroupSize(limit int32) string {
	return fmt.Sprintf("no se pueden superar los %v elementos", limit)
}

func GetSkillRegisterInput(r *http.Request) (SkillRegisterInput, SkillRegisterErrors, bool) {
	name := r.FormValue("name")
	errFound := false
	inputErrors := SkillRegisterErrors{}
	if len(name) < 1 || len(name) > 100 {
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
	CountSkills(ctx context.Context, userID string) (int32, error)
	RegisterSkill(ctx context.Context, skillID, userID, name string) error
	DeleteSkill(ctx context.Context, userID, skillID string) error
}

type registerSkillInputFn func(r *http.Request) (SkillRegisterInput, SkillRegisterErrors, bool)

func CreateRegisterSkillHandler(templ shared.TemplatesRepo, authz shared.AuthRep, storage SkillStorage, inputFn registerSkillInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authz.GetUser(r)
		if err != nil || user.Role == auth.NotAuthRole {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		count, err := storage.CountSkills(r.Context(), user.ID)
		if err != nil || count >= maxGroupSize {
			inputErrors.NameError = errGroupSize(maxGroupSize)
		} else {
			if err := storage.RegisterSkill(r.Context(), skillID, user.ID, input.Name); err != nil {
				inputErrors.NameError = errUnexpected
			}
		}
		if inputErrors.NameError != "" {
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
