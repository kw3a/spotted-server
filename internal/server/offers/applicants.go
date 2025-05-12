package offers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferApplInput struct {
	OfferID string
}

type OfferApplData struct {
	User       auth.AuthUser
	Offer      shared.Offer
	Quiz       shared.Quiz
	Problems   []shared.Problem
	Languages  []shared.Language
	Applicants []shared.Application
}

type OfferApplStorage interface {
	SelectOfferByUser(ctx context.Context, id string, userID string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
	SelectApplications(ctx context.Context, quizID string) ([]shared.Application, error)
	SelectFullProblems(ctx context.Context, quizID string) ([]shared.Problem, error)
}

func GetOfferApplInput(r *http.Request) (OfferApplInput, error) {
	offerID := chi.URLParam(r, "offerID")
	if err := shared.ValidateUUID(offerID); err != nil {
		return OfferApplInput{}, err
	}
	return OfferApplInput{OfferID: offerID}, nil
}

type offerApplInputFn func(r *http.Request) (OfferApplInput, error)

func CreateOfferApplHandler(
	inputFn offerApplInputFn,
	authService shared.AuthRep,
	storage OfferApplStorage,
	templ shared.TemplatesRepo,
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
		offer, err := storage.SelectOfferByUser(r.Context(), input.OfferID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quiz, err := storage.SelectQuizByOffer(r.Context(), input.OfferID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		languages, err := storage.SelectLanguages(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		applications, err := storage.SelectApplications(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		problems, err := storage.SelectFullProblems(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := OfferApplData{
			User:       user,
			Offer:      offer,
			Quiz:       quiz,
			Problems:   problems,
			Languages:  languages,
			Applicants: applications,
		}
		if err := templ.Render(w, "offerAdmin", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
