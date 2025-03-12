package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/offers"
)

func (DI *App) JobOffersHandler() http.HandlerFunc {
	return offers.CreateJobOffersHandler(
		offers.GetJobOffersParams,
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}

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

func (DI *App) OffersAdmin() http.HandlerFunc {
	return offers.CreateOffersAdminHandler(
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}

func (DI *App) OfferArchive() http.HandlerFunc {
	return offers.CreateOfferArchiveHandler(
		offers.GetOfferArchiveInput,
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}

func (DI *App) OfferAdmin() http.HandlerFunc {
	return offers.CreateOfferApplHandler(
		offers.GetOfferApplInput,
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}
