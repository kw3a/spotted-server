package offers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errNotAuthorized = "you must be authenticated to use this function"
)

type OfferArchiveStorage interface {
	ArchiveOffer(ctx context.Context, offerID, ownerID string) error
}

type OfferArchiveInput struct {
	OfferID string
}

func GetOfferArchiveInput(r *http.Request) (OfferArchiveInput, error) {
	offerID := chi.URLParam(r, "offerID")
	if err := shared.ValidateUUID(offerID); err != nil {
		return OfferArchiveInput{}, err
	}
	return OfferArchiveInput{OfferID: offerID}, nil
}


type offerArchiveInputFn func(r *http.Request) (OfferArchiveInput, error)
func CreateOfferArchiveHandler(
	inputFn offerArchiveInputFn,
	authService shared.AuthRep,
	storage OfferArchiveStorage,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if user.Role == "visitor" || user.ID == "" {
			http.Error(w, errNotAuthorized, http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = storage.ArchiveOffer(r.Context(), input.OfferID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
