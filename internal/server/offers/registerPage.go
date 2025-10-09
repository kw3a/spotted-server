package offers

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type RegisterOfferStorage interface {
	GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error)
	GetLanguages(ctx context.Context) ([]shared.Language, error)
}

type RegisterOfferData struct {
	User      auth.AuthUser
	Companies []shared.Company
	Offers    []shared.Offer
	Languages []shared.Language
}

func CreateRegisterPage(
	authz shared.AuthRep,
	templ shared.TemplatesRepo,
	storage RegisterOfferStorage,
	redirection string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authz.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if user.Role == auth.NotAuthRole {
			http.Redirect(w, r, redirection, http.StatusSeeOther)
			return
		}
		params := shared.CompanyQueryParams{UserID: user.ID, Page: 1}
		companies, err := storage.GetCompanies(r.Context(), params)
		if err != nil || len(companies) == 0 {
			http.Redirect(w, r, redirection, http.StatusSeeOther)
			return
		}
		languages, err := storage.GetLanguages(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := RegisterOfferData{
			User:      user,
			Companies: companies,
			Languages: languages,
		}
		if err := templ.Render(w, "offerPage", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
