package offers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferEditionPageData struct {
	User      auth.AuthUser
	Offer     shared.Offer
	Quiz      shared.Quiz
	Languages []shared.Language
}

type OfferEditionPageInput struct {
	OfferID string
}

type OfferEditionPageStorage interface {
	SelectOfferByUser(ctx context.Context, ID, userID string) (shared.Offer, error)
	GetLanguages(ctx context.Context) ([]shared.Language, error)
	SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error)
}

func GetOfferEditionPageInput(r *http.Request) (OfferEditionPageInput, error) {
	offerID := chi.URLParam(r, "offerID")
	if err := shared.ValidateUUID(offerID); err != nil {
		return OfferEditionPageInput{}, err
	}
	return OfferEditionPageInput{OfferID: offerID}, nil
}

type offerEditionPageInputFn func(r *http.Request) (OfferEditionPageInput, error)

func CreateOfferEditionPage(
	auth shared.AuthRep,
	templ shared.TemplatesRepo,
	storage OfferEditionPageStorage,
	inputFn offerEditionPageInputFn,
) http.HandlerFunc {
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
		offer, err := storage.SelectOfferByUser(r.Context(), input.OfferID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		languages, err := storage.GetLanguages(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := OfferEditionPageData{
			User:      user,
			Offer:     offer,
			Languages: languages,
		}
		quiz, err := storage.SelectQuizByOffer(r.Context(), offer.ID)
		if err == nil {
			data.Quiz = quiz
		}
		if err := templ.Render(w, "offerEdition", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
