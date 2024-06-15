package auth

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/http_parser/bodyParams"
	res "github.com/kw3a/spotted-server/internal/http_parser/responseParser"
)

func CreateRevokeHandler(authRep Authentication, authStorage AuthenticationStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := bodyParams.RevokeParams{}
		params, err := body.Parse(r)
		if err != nil {
			res.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err = authRep.ValidateRefresh(params.RefreshToken); err != nil {
			res.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err = authStorage.Revoke(r.Context(), params.RefreshToken); err != nil {
			res.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		res.Ok(w)
	}
}

func (authService *AuthService) RevokeHandler(jwtSecret string, db *database.Queries) http.HandlerFunc {
	return CreateRevokeHandler(
		authService.JWTRep,
		authService.Storage,
	)
}
