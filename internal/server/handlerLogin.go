package server

import (
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/kw3a/spotted-server/internal/auth"
)

func (apiCfg *ApiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type paramaters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type responseBody struct {
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := paramaters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse parameters")
		return
	}

	dbUser, err := apiCfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Email not found")
		return
	}
	err = auth.CheckPasswordHash(params.Password, dbUser.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid password")
		return
	}
	token, err := auth.MakeJWT(dbUser.ID, apiCfg.jwtSecret, time.Hour*2, auth.TokenTypeAccess)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, responseBody{
		Token: token,
	})
}
