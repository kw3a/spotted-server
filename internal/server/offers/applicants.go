package offers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ApplicantsInput struct {
	OfferID string
}

type ApplicantsData struct {
	VideoRecordsURL string
	User            auth.AuthUser
	Offer           shared.Offer
	Quiz            shared.Quiz
	Problems        []shared.Problem
	Languages       []shared.Language
	Applicants      []ApplicantReview
}

type ApplicantsStorage interface {
	SelectOfferByUser(ctx context.Context, id string, userID string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
	SelectApplications(ctx context.Context, quizID string) ([]shared.Application, error)
	SelectFullProblems(ctx context.Context, quizID string) ([]shared.Problem, error)
}

func GetApplicantsInput(r *http.Request) (ApplicantsInput, error) {
	offerID := chi.URLParam(r, "offerID")
	if err := shared.ValidateUUID(offerID); err != nil {
		return ApplicantsInput{}, err
	}
	return ApplicantsInput{OfferID: offerID}, nil
}

type offerApplInputFn func(r *http.Request) (ApplicantsInput, error)

func redirectApplicantsToHome(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateApplicantsHandler(
	inputFn offerApplInputFn,
	authService shared.AuthRep,
	storage ApplicantsStorage,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			redirectApplicantsToHome(w, r)
			return
		}
		if user.Role == auth.NotAuthRole {
			redirectApplicantsToHome(w, r)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		offer, err := storage.SelectOfferByUser(r.Context(), input.OfferID, user.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				redirectApplicantsToHome(w, r)
				return
			}
			http.Error(w, "could not load offer", http.StatusInternalServerError)
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
		videoRecordsURL := os.Getenv("VIDEO_RECORDS_URL")
		data := ApplicantsData{
			VideoRecordsURL: videoRecordsURL,
			User:            user,
			Offer:           offer,
			Quiz:            quiz,
			Problems:        problems,
			Languages:       languages,
			Applicants:      buildApplicantReviews(applications, problems),
		}
		if err := templ.Render(w, "offerAdmin", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
