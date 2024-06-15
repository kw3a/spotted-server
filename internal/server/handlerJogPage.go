package server

import (
	"context"
	"net/http"
)

type JobPageData struct {
	Offers []Offer
}

type Offer struct {
	ID          string
	Title       string
	Description string
}

type JobOfferStorage interface {
	SelectOffers(ctx context.Context) ([]Offer, error)
}

func createJobOffersHandler(templ *Templates, storage JobOfferStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		offers, err := storage.SelectOffers(r.Context())
		if err != nil {
			http.Error(w, "can't find offers", http.StatusInternalServerError)
			return
		}
		data := JobPageData{Offers: offers}
		err = templ.Render(w, "jobPage", data)
		if err != nil {
			http.Error(w, "can't render job page", http.StatusInternalServerError)
		}
	}
}

func (DI *App) JobOffersHandler() http.HandlerFunc {
	return createJobOffersHandler(DI.Templ, DI.Storage)
}
