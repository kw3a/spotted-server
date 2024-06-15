package auth

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/http_parser/bodyParams"
	res "github.com/kw3a/spotted-server/internal/http_parser/responseParser"
)

func CreateRefreshHandler(authRep Authentication, authStorage AuthenticationStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := bodyParams.RefreshParams{}
		params, err := body.Parse(r)
		if err != nil {
			res.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if err := authRep.ValidateRefresh(params.RefreshToken); err != nil {
			res.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if err := authStorage.IsRegistered(r.Context(), params.RefreshToken); err != nil {
			res.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		accessToken, err := authRep.CreateAccess(params.RefreshToken)
		if err != nil {
			res.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		responseBody := res.GetRefreshResponseBody(accessToken)
		responseBody.Send(w)
	}
}

func (authService *AuthService) RefreshHandler(jwtSecret string, db *database.Queries) http.HandlerFunc {
	return CreateRefreshHandler(
		authService.JWTRep,
		authService.Storage,
	)
}
