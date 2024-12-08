package companies

import (
	"context"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type CompanyListStorage interface {
	GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error)
}

type CompanyListData struct {
	User      auth.AuthUser
	Companies []shared.Company
}


func GetCompanyListParams(r *http.Request) shared.CompanyQueryParams {
	res := shared.CompanyQueryParams{}
	q := r.URL.Query()
	userID := q.Get("u")
	if shared.ValidateUUID(userID) == nil {
		res.UserID = userID
	}
	query := q.Get("q")	
	if query != "" {
		res.Query = query
	}
	return res
}

type companyListParamsFn func(r *http.Request) shared.CompanyQueryParams

func CreateCompanyListPageHandler(
	authService shared.AuthRep,
	templ shared.TemplatesRepo,
	storage CompanyListStorage,
	paramsFn companyListParamsFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		params := paramsFn(r)
		companies, err := storage.GetCompanies(r.Context(), params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data := CompanyListData{
			User:      user,
			Companies: companies,
		}
		if err := templ.Render(w, "companyList", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
