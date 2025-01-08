package offers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferRegError struct {
	Error string
}

type OfferStorage interface {
	RegisterOffer(
		ctx context.Context,
		offerID string, offer shared.Offer,
		quizID string, quiz shared.Quiz,
		problems []shared.Problem,
	) error
	GetCompanyByID(ctx context.Context, companyID string) (shared.Company, error)
}

type OfferRegInputFn func(r *http.Request) (OfferRegInput, error)

func CreateOfferRegistrationHandler(
	templ shared.TemplatesRepo,
	auth shared.AuthRep,
	storage OfferStorage,
	redirPath string,
	inputFn OfferRegInputFn,
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
		company, err := storage.GetCompanyByID(r.Context(), input.Offer.CompanyID)
		if err != nil {
			http.Error(w, fmt.Sprintf("storage error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		if company.UserID != user.ID {
			http.Error(w, "you are not the owner of this company", http.StatusUnauthorized)
			return
		}
		offerID := uuid.New().String()
		quizID := uuid.New().String()
		if err := storage.RegisterOffer(
			r.Context(),
			offerID,
			input.Offer,
			quizID,
			input.Quiz,
			input.Problems,
		); err != nil {
			http.Error(w, fmt.Sprintf("storage error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Add("HX-Redirect", redirPath+offerID)
		w.WriteHeader(http.StatusOK)
	}
}
