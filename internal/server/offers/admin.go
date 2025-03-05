package offers

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OffersAdminData struct {
	User   auth.AuthUser
	Offers []shared.Offer
}

type OffersAdminStorage interface {
	SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error)
}

func CreateOffersAdminHandler(
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
		if user.Role == "visitor" {
			if err := templ.Render(w, "offersAdmin", OffersAdminData{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		offers, err := storage.SelectOffers(r.Context(), shared.OfferQueryParams{UserID: user.ID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := OffersAdminData{
			User:   user,
			Offers: offers,
		}
		if err := templ.Render(w, "offersAdmin", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
