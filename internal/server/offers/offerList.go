package offers

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferListData struct {
	User   auth.AuthUser
	Offers []shared.Offer
}

type OfferListStorage interface {
	SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error)
}

func GetJobOffersParams(r *http.Request) shared.OfferQueryParams {
	params := shared.OfferQueryParams{}
	q := r.URL.Query()
	query := q.Get("q")
	if query != "" {
		params.Query = query
	}
	user := q.Get("u")
	if shared.ValidateUUID(user) == nil {
		params.UserID = user
	}
	return params
}

type offerListParamsFn func(r *http.Request) shared.OfferQueryParams

func CreateJobOffersHandler(
	paramsFn offerListParamsFn,
	authService shared.AuthRep,
	storage OfferListStorage,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		params := paramsFn(r)
		offers, err := storage.SelectOffers(r.Context(), params)
		if err != nil {
			http.Error(w, "can't find offers", http.StatusInternalServerError)
			return
		}
		data := OfferListData{
			User:   user,
			Offers: offers,
		}
		if err = templ.Render(w, "jobPage", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
