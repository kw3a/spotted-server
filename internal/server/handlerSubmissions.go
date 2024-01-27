package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/kw3a/spotted-server/internal/database"
)

func (apiCfg *ApiConfig) handlerSubmissionCreate(w http.ResponseWriter, r *http.Request, dbUser database.User) {
	quizID := chi.URLParam(r, "quizID")
	err := uuid.Validate(quizID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid quiz id")
		return
	}
	problemID := chi.URLParam(r, "problemID")
	err = uuid.Validate(problemID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid problem id")
		return
	}
	type paramaters struct {
		Src        string `json:"src"`
		LanguageID int32  `json:"language_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := paramaters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse parameters")
		return
	}

	err = apiCfg.DB.CreateSubmission(r.Context(), database.CreateSubmissionParams{
		ID:         uuid.New().String(),
		Src:        params.Src,
		Time:       time.Now(),
		ProblemID:  problemID,
		UserID:     dbUser.ID,
		LanguageID: params.LanguageID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, nil)
}
