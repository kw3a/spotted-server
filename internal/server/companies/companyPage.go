package companies

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type CompanyPageData struct {
	User    auth.AuthUser
	Company shared.Company
	Offers  []shared.Offer
}

type CompanyPageInput struct {
	CompanyID string
}

type CompanyPageStorage interface {
	GetCompanyByID(ctx context.Context, companyID string) (shared.Company, error)
	SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error)
}

func GetCompanyPageInput(r *http.Request) (CompanyPageInput, error) {
	companyID := chi.URLParam(r, "companyID")
	if err := shared.ValidateUUID(companyID); err != nil {
		return CompanyPageInput{}, err
	}
	return CompanyPageInput{
		CompanyID: companyID,
	}, nil
}

type companyPageInputFn func(r *http.Request) (CompanyPageInput, error)

func CreateCompanyPageHandler(
	templ shared.TemplatesRepo,
	authService shared.AuthRep,
	storage CompanyPageStorage,
	inputFn companyPageInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		company, err := storage.GetCompanyByID(r.Context(), input.CompanyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query := shared.OfferQueryParams{CompanyID: input.CompanyID, Page: shared.PageParam(r)}
		offers, err := storage.SelectOffers(r.Context(), query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := CompanyPageData{
			User:    user,
			Company: company,
			Offers:  offers,
		}
		err = templ.Render(w, "companyPage", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
