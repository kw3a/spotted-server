package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/companies"
)

func (DI *App) CompanyListPageHandler() http.HandlerFunc {
	return companies.CreateCompanyListPageHandler(
		DI.AuthService,
		DI.Templ,
		DI.Storage,
		companies.GetCompanyListParams,
	)
}

func (DI *App) CompanyRegistrationPageHandler() http.HandlerFunc {
	return companies.CreateRegisterCompanyPage(
		DI.Templ,
		DI.AuthService,
	)
}

func (DI *App) CompanyRegistrationHandler() http.HandlerFunc {
	return companies.CreateRegisterCompanyHandler(
		DI.Storage,
		DI.AuthService,
		&DI.Cld.Upload,
		companies.GetRegisterCompanyInput,
		DI.Templ,
	)
}

func (DI *App) CompanyPageHandler() http.HandlerFunc {
	return companies.CreateCompanyPageHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		companies.GetCompanyPageInput,
	)
}
