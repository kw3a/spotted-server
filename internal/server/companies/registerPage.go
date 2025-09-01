package companies

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type RegisterCompanyPageData struct {
	User         auth.AuthUser
}

func CreateRegisterPageHandler(templ shared.TemplatesRepo, authService shared.AuthRep, redirectPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if user.Role == auth.NotAuthRole {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
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
