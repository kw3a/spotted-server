package companies

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type RegisterCompanyPageData struct {
	User         auth.AuthUser
	DefaultImage string
}

func CreateRegisterCompanyPage(templ shared.TemplatesRepo, authService shared.AuthRep, defaultImgPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := RegisterCompanyPageData{
			User: user,
			DefaultImage: defaultImgPath,
		}
		err = templ.Render(w, "companyRegistration", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
