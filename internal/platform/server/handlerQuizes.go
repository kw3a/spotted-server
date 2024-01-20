package server

import (
	"net/http"
)

func (apiCfg *ApiConfig) handlerQuizzesGet(w http.ResponseWriter, r *http.Request) {
	dbQuizzes, err := apiCfg.DB.GetQuizzes(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Fatal(respondWithJSON(w, http.StatusOK, (dbQuizzes)))
}
