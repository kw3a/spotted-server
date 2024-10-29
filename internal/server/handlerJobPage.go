package server

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
)

type JobPageData struct {
	User   auth.AuthUser
	Offers []PartialOffer
}

type PartialOffer struct {
	QuizID       string
	Title        string
	Author       string
	MinWage      int32
	MaxWage      int32
	RelativeTime string
}

type JobOfferStorage interface {
	SelectOffers(ctx context.Context) ([]PartialOffer, error)
}

func CreateJobOffersHandler(authService AuthRep, templ TemplatesRepo, storage JobOfferStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		offers, err := storage.SelectOffers(r.Context())
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
	return CreateJobOffersHandler(DI.AuthService, DI.Templ, DI.Storage)
}
