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
	CmpSearch string
	CmpUser   string
	NextPage  int32
}

func GetCompanyListParams(r *http.Request) shared.CompanyQueryParams {
	res := shared.CompanyQueryParams{}
	q := r.URL.Query()
	userID := q.Get("u")
	if shared.ValidateUUID(userID) == nil {
		res.UserID = userID
	}
	query := q.Get("q")
	if len(query) > 2 && len(query) < 65 {
		res.Query = query
	}
	res.Page = shared.PageParam(r)
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
			NextPage:  params.Page + 1,
		}
		if params.UserID != "" {
			data.CmpUser = params.UserID
		}
		if params.Query != "" {
			data.CmpSearch = params.Query
		}
		toRender := "companyListPage"
		if params.Page > 1 {
			toRender = "companyList"
		}
		if err := templ.Render(w, toRender, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
