package profiles

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ProfilePageData struct {
	User           auth.AuthUser
	Profile        Profile
	Links          []shared.Link
	Experiences    []shared.ExperienceEntry
	Education      []shared.EducationEntry
	Skills         []shared.SkillEntry
	Participations []shared.Offer
}
type Profile struct {
	Name        string
	ImageURL    string
	Description string
}
type ProfilePageInput struct {
	UserID string
}

func GetProfilePageInput(r *http.Request) (ProfilePageInput, error) {
	userID := chi.URLParam(r, "userID")
	err := shared.ValidateUUID(userID)
	if err != nil {
		return ProfilePageInput{}, err
	}
	return ProfilePageInput{
		UserID: userID,
	}, nil
}

type ProfilePageStorage interface {
	GetUser(ctx context.Context, userID string) (database.User, error)
	SelectExperiences(ctx context.Context, userID string) ([]shared.ExperienceEntry, error)
	SelectEducation(ctx context.Context, userID string) ([]shared.EducationEntry, error)
	SelectSkills(ctx context.Context, userID string) ([]shared.SkillEntry, error)
	SelectLinks(ctx context.Context, userID string) ([]shared.Link, error)
	SelectParticipatedOffers(ctx context.Context, userID string) ([]shared.Offer, error)
}

type profilePageInputFunc func(r *http.Request) (ProfilePageInput, error)

func CreateProfilePageHandler(
	authService shared.AuthRep,
	templ shared.TemplatesRepo,
	storage ProfilePageStorage,
	inputFn profilePageInputFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dbUser, err := storage.GetUser(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		experiences, err := storage.SelectExperiences(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		education, err := storage.SelectEducation(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		skills, err := storage.SelectSkills(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		links, err := storage.SelectLinks(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		participatedOffers, err := storage.SelectParticipatedOffers(r.Context(), input.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		profile := Profile{
			Name:        dbUser.Name,
			ImageURL:    dbUser.ImageUrl,
			Description: dbUser.Description,
		}
		data := ProfilePageData{
			User:           user,
			Profile:        profile,
			Links:          links,
			Experiences:    experiences,
			Education:      education,
			Skills:         skills,
			Participations: participatedOffers,
		}
		if err = templ.Render(w, "profilePage", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
