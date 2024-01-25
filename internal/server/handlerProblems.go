package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/kw3a/spotted-server/internal/database"
)

func (apiCfg *ApiConfig) handlerProblemGet(w http.ResponseWriter, r *http.Request) {
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
	dbProblem, err := apiCfg.DB.GetProblem(r.Context(), database.GetProblemParams{
		QuizID:   quizID,
		ID:       problemID,
		QuizID_2: quizID,
		ID_2:     problemID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, databaseProblemToProblem(dbProblem))
}

func (apiCfg *ApiConfig) handlerProblemsGet(w http.ResponseWriter, r *http.Request) {
	quizID := chi.URLParam(r, "quizID")
	err := uuid.Validate(quizID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid quiz id")
		return
	}
	dbProblems, err := apiCfg.DB.GetProblems(r.Context(), database.GetProblemsParams{
		QuizID:   quizID,
		QuizID_2: quizID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(dbProblems) == 0 {
		respondEmpty(w, http.StatusOK, "Problems not found")
		return
	}
	respondWithJSON(w, http.StatusOK, databaseProblemsToProblems(dbProblems))
}

func (apiCfg *ApiConfig) handlerProblemCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Description string  `json:"description"`
		Title       string  `json:"title"`
		MemoryLimit int32   `json:"memoryLimit"`
		TimeLimit   float64 `json:"timeLimit"`
		QuizID      string  `json:"quiz_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse parameters")
		return
	}
	err = apiCfg.DB.CreateProblem(r.Context(), database.CreateProblemParams{
		ID:          uuid.New().String(),
		Description: params.Description,
		Title:       params.Title,
		MemoryLimit: params.MemoryLimit,
		TimeLimit:   params.TimeLimit,
		QuizID:      params.QuizID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't create")
		return
	}
	w.WriteHeader(http.StatusCreated)
}
