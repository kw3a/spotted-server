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
	Problems   []*shared.Problem
	Languages  []shared.Language
	Applicants []Application
}

type Application struct {
	Applicant     auth.AuthUser
	Participation shared.Participation
	Summary       []Summary
}

type Summary struct {
	Title      string
	Score      shared.Score
	Submission shared.Submission
	Results    []shared.TestCaseResult
}

type OfferApplStorage interface {
	SelectOfferByUser(ctx context.Context, id string, userID string) (shared.Offer, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
	SelectApplicants(ctx context.Context, quizID string) ([]auth.AuthUser, error)
	ParticipationStatus(ctx context.Context, userID string, quizID string) (shared.Participation, error)
	SelectProblems(ctx context.Context, quizID string) ([]shared.Problem, error)
	SelectScore(ctx context.Context, userID string, problemID string) (shared.Score, error)
	SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error)
	SelectExamples(ctx context.Context, problemID string) ([]shared.Example, error)
	SelectTestCases(ctx context.Context, problemID string) ([]shared.TestCase, error)
	BestSubmission(ctx context.Context, applicantID, problemID string) (shared.Submission, error)
	GetResults(ctx context.Context, problemID, submissionID string) ([]shared.TestCaseResult, error)
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
		applicants, err := storage.SelectApplicants(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		problems, err := storage.SelectProblems(r.Context(), quiz.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pointers := make([]*shared.Problem, 0, len(problems))
		for i := range problems {
			pointers = append(pointers, &problems[i])
		}
		for _, problem := range pointers {
			examples, err := storage.SelectExamples(r.Context(), problem.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			test_cases, err := storage.SelectTestCases(r.Context(), problem.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			problem.Examples = examples
			problem.TestCases = test_cases
		}
		data := OfferApplData{
			User:      user,
			Offer:     offer,
			Quiz:      quiz,
			Problems:  pointers,
			Languages: languages,
		}

		for _, applicant := range applicants {
			participation, err := storage.ParticipationStatus(r.Context(), applicant.ID, quiz.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			apl := Application{
				Applicant:     applicant,
				Participation: participation,
			}
			for _, problem := range problems {
				score, err := storage.SelectScore(r.Context(), applicant.ID, problem.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				subm, err := storage.BestSubmission(r.Context(), applicant.ID, problem.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				results, err := storage.GetResults(r.Context(), problem.ID, subm.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				apl.Summary = append(apl.Summary, Summary{
					Title:      problem.Title,
					Score:      score,
					Submission: subm,
					Results:    results,
				})
			}
			data.Applicants = append(data.Applicants, apl)
		}
		if err := templ.Render(w, "offerAdmin", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
