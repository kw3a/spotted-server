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
	Applicants []Application
}

type Application struct {
	Applicant auth.AuthUser
	Results   []Result
}

type Result struct {
	Problem shared.Problem
	Score   shared.Score
}

type OfferApplStorage interface {
	SelectOfferByUser(ctx context.Context, id string, userID string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectApplicants(ctx context.Context, quizID string) ([]auth.AuthUser, error)
	SelectProblems(ctx context.Context, quizID string) ([]shared.Problem, error)
	SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error)
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
		problems, err := storage.SelectProblems(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		applicants, err := storage.SelectApplicants(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := OfferApplData{
			User:  user,
			Offer: offer,
		}

		for _, applicant := range applicants {
			app := Application{Applicant: applicant}
			for _, problem := range problems {
				score, err := storage.SelectScore(r.Context(), applicant.ID, problem.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				app.Results = append(app.Results, Result{
					Problem: problem,
					Score: score,
				})
			}
			data.Applicants = append(data.Applicants, app)
		}
		if err := templ.Render(w, "offerAdmin", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
