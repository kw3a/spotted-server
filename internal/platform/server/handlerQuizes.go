package server

import (
	"log"
	"net/http"
)

func (apiCfg *ApiConfig) handlerQuizzesGet(w http.ResponseWriter, r *http.Request) {
	dbQuizzes, err := apiCfg.DB.GetQuizzes(r.Context())
	if err != nil {
		log.Fatal(respondWithError(w, http.StatusInternalServerError, err.Error()))
	}
	log.Fatal(respondWithJSON(w, http.StatusOK, (dbQuizzes)))

}
