package profiles

import (
	"context"
	"log"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
)

type logoutStorage interface {
	Revoke(ctx context.Context, refreshToken string) error
}
func CreateLogoutHandler(storage logoutStorage, redirectPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := r.Cookie("refresh_token")
		if err == nil {
			if err = storage.Revoke(r.Context(), refreshToken.Value); err != nil {
				log.Println(err)
			}
		}
		auth.DeleteCookies(w)
		w.Header().Set("HX-Redirect", redirectPath)
	}
}
