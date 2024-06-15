package responseparser

import (
	"encoding/json"
	"log"
	"net/http"
)

type EmptyResponse struct {
	Message string `json:"error_message"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondEmpty(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, EmptyResponse{
		Message: msg,
	})
}
