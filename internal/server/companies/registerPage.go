package companies

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type RegisterCompanyPageData struct {
	User         auth.AuthUser
}

func CreateRegisterCompanyPage(templ shared.TemplatesRepo, authService shared.AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := RegisterCompanyPageData{
			User: user,
		}
		err = templ.Render(w, "companyRegistration", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
