package profiles

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type UserFormData struct {
	User auth.AuthUser
}

func CreateRegPageHandler(authService shared.AuthRep, templ shared.TemplatesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		err = templ.Render(w, "userCreation", UserFormData{
			User: user,
		})
		if err != nil {
			http.Error(w, "can't render user page", http.StatusInternalServerError)
		}
	}
}
