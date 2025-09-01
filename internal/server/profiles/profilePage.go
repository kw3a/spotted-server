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
	Profile        auth.AuthUser
	Links          []shared.Link
	Experiences    []shared.ExperienceEntry
	Education      []shared.EducationEntry
	Skills         []shared.SkillEntry
	Participations []shared.Offer
	NextPage       int32
	Owner          bool
}
type ProfilePageInput struct {
	UserID string
	Page   int32
}

func GetProfilePageInput(r *http.Request) (ProfilePageInput, error) {
	userID := chi.URLParam(r, "userID")
	err := shared.ValidateUUID(userID)
	if err != nil {
		return ProfilePageInput{}, err
	}
	return ProfilePageInput{
		UserID: userID,
		Page:   shared.PageParam(r),
	}, nil
}

type ProfilePageStorage interface {
	GetUser(ctx context.Context, userID string) (database.User, error)
	SelectExperiences(ctx context.Context, userID string) ([]shared.ExperienceEntry, error)
	SelectEducation(ctx context.Context, userID string) ([]shared.EducationEntry, error)
	SelectSkills(ctx context.Context, userID string) ([]shared.SkillEntry, error)
	SelectLinks(ctx context.Context, userID string) ([]shared.Link, error)
	SelectParticipatedOffers(ctx context.Context, userID string, page int32) ([]shared.Offer, error)
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
		participatedOffers, err := storage.SelectParticipatedOffers(r.Context(), input.UserID, input.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		toRender := "profilePage"
		data := ProfilePageData{}
		data.NextPage = input.Page + 1
		if input.Page > 1 {
			toRender = "participationList"
			data.Participations = participatedOffers
		} else {
			dbProfile, err := storage.GetUser(r.Context(), input.UserID)
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
			profile := auth.AuthUser{
				ID:          dbProfile.ID,
				Name:        dbProfile.Name,
				ImageURL:    dbProfile.ImageUrl,
				Description: dbProfile.Description,
				Email:       dbProfile.Email,
				Cell:        dbProfile.Number,
			}
			data = ProfilePageData{
				User:           user,
				Profile:        profile,
				Links:          links,
				Experiences:    experiences,
				Education:      education,
				Skills:         skills,
				Participations: participatedOffers,
				Owner:          profile.ID == user.ID,
				NextPage:       input.Page + 1,
			}
		}
		if err = templ.Render(w, toRender, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
