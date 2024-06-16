package auth

import (
	"log"
	"net/http"

	"github.com/kw3a/spotted-server/internal/database"
)

func deleteCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func createLogoutHandler(storage AuthenticationStorage, redirectPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := r.Cookie("refresh_token")
		if err != nil {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}
		if err = storage.Revoke(r.Context(), refreshToken.Value); err != nil {
			log.Println(err)
		}
		deleteCookies(w)
		http.Redirect(w, r, redirectPath, http.StatusSeeOther)
	}
}

func (authService *AuthService) LogoutHandler(jwtSecret string, db *database.Queries) http.HandlerFunc {
	return createLogoutHandler(authService.Storage, "/")
}
