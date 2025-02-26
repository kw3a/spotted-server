package server

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type JobPageData struct {
	User   auth.AuthUser
	Offers []shared.Offer
}

type JobOfferStorage interface {
	SelectOffers(ctx context.Context, params shared.JobQueryParams) ([]shared.Offer, error)
}

func GetJobOffersParams(r *http.Request) shared.JobQueryParams {
	res := shared.JobQueryParams{}
	q := r.URL.Query()
	query := q.Get("q")
	if query != "" {
		res.Query = query
	}
	return res
}

type jobOffersParamsFn func(r *http.Request) shared.JobQueryParams
func CreateJobOffersHandler(
	authService AuthRep,
	templ TemplatesRepo,
	storage JobOfferStorage,
	paramsFn jobOffersParamsFn,
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
		data := JobPageData{
			User:   user,
			Offers: offers,
		}
		if err = templ.Render(w, "jobPage", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (DI *App) JobOffersHandler() http.HandlerFunc {
	return CreateJobOffersHandler(DI.AuthService, DI.Templ, DI.Storage, GetJobOffersParams)
}
