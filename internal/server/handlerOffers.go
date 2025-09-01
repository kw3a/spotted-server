package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/offers"
)

func (DI *App) JobOffersHandler() http.HandlerFunc {
	return offers.CreateOfferListHandler(
		offers.GetJobOffersParams,
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}

func (app *App) PreambleHandler() http.HandlerFunc {
	return offers.CreateParticipationHandler(
		app.Templ,
		app.Storage,
		app.AuthService,
		offers.GetPreambleInput,
	)
}

func (DI *App) OfferRegistrationPage() http.HandlerFunc {
	return offers.CreateRegisterPage(
		DI.AuthService,
		DI.Templ,
		DI.Storage,
		"/register/companies",
	)
}

func (DI *App) OfferRegistration() http.HandlerFunc {
	return offers.CreateRegisterHandler(
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
	return offers.CreateArchiveHandler(
		offers.GetOfferArchiveInput,
		DI.AuthService,
		DI.Storage,
	)
}

func (DI *App) OfferAdmin() http.HandlerFunc {
	return offers.CreateApplicantsHandler(
		offers.GetApplicantsInput,
		DI.AuthService,
		DI.Storage,
		DI.Templ,
	)
}
