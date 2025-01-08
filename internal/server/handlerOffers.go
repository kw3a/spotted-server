package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/offers"
)

func (DI *App) OfferRegistrationPage() http.HandlerFunc {
	return offers.CreateRegisterOfferPage(
		DI.AuthService,
		DI.Templ,
		DI.Storage,
		"/register/companies",
	)
}

func (DI *App) OfferRegistration() http.HandlerFunc {
	return offers.CreateOfferRegistrationHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		"/preamble/",
		offers.GetOfferRegInput,
	)
}

func (DI *App) OfferEditionPage() http.HandlerFunc {
	return offers.CreateOfferEditionPage(
		DI.AuthService,
		DI.Templ,
		DI.Storage,
		offers.GetOfferEditionPageInput,
	)
}

func (DI *App) OfferEdition() http.HandlerFunc {
	return offers.CreateOfferEdition(
		DI.Storage,
		DI.Templ,
		offers.GetOfferEditionInput,
	)
}
