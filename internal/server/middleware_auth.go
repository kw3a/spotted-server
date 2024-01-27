package server

import (
	"net/http"

	"gitlab.com/kw3a/spotted-server/internal/auth"
	"gitlab.com/kw3a/spotted-server/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *ApiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		userID, err := auth.ValidateJWT(token, apiCfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		dbUser, err := apiCfg.DB.GetUser(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		handler(w, r, dbUser)
	}
}
