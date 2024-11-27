package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/database"
)

type ProfilePageData struct {
	User        auth.AuthUser
	Profile     Profile
	Links       []Link
	Experiences []ExperienceEntry
	Education   []EducationEntry
	Skills      []SkillEntry
}
type Profile struct {
	Name        string
	ImageURL    string
	Description string
}
type Link struct {
	URL  string
	Name string
	ID   string
}
type ExperienceEntry struct {
	ID          string
	Title       string
	Company     string
	StartDate   string
	EndDate     string
	Description string
}
type EducationEntry struct {
	ID					string
	Degree      string
	Institution string
	StartDate   string
	EndDate     string
}

type ProfilePageInput struct {
	UserID string
}

func GetProfilePageInput(r *http.Request) (ProfilePageInput, error) {
	userID := chi.URLParam(r, "userID")
	err := ValidateUUID(userID)
	if err != nil {
		return ProfilePageInput{}, err
	}
	return ProfilePageInput{
		UserID: userID,
	}, nil
}

type ProfilePageStorage interface {
	GetUser(ctx context.Context, userID string) (database.User, error)
	SelectExperiences(ctx context.Context, userID string) ([]ExperienceEntry, error)
	SelectEducation(ctx context.Context, userID string) ([]EducationEntry, error)
	SelectSkills(ctx context.Context, userID string) ([]SkillEntry, error)
	SelectLinks(ctx context.Context, userID string) ([]Link, error)
}

type profilePageInputFunc func(r *http.Request) (ProfilePageInput, error)

func CreateProfilePageHandler(
	authService AuthRep,
	templ TemplatesRepo,
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
		imageURL := dbUser.ImageUrl
		if imageURL == "" {
			imageURL = defaultImagePath
		}
		profile := Profile{
			Name:        dbUser.Name,
			ImageURL:    imageURL,
			Description: dbUser.Description,
		}
		data := ProfilePageData{
			User:        user,
			Profile:     profile,
			Links:       links,
			Experiences: experiences,
			Education:   education,
			Skills:      skills,
		}
		if err = templ.Render(w, "profilePage", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (DI *App) ProfilePageHandler() http.HandlerFunc {
	return CreateProfilePageHandler(DI.AuthService, DI.Templ, DI.Storage, GetProfilePageInput)
}
