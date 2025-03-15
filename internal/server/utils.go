package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/auth"
)

type AuthRep interface {
	GetUser(r *http.Request) (userID auth.AuthUser, err error)
}
func ValidateUUID(id string) error {
	return uuid.Validate(id)
}
type ErrorMesage struct {
	Msg string `json:"msg"`
}



func Encode(w http.ResponseWriter, status int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func ErrorLog(err error) {
	if err != nil {
		log.Println(err)
	}
}
func EncodeAndLog[T any](w http.ResponseWriter, status int, v T) {
	err := Encode(w, status, v)
	if err != nil {
		log.Println(err)
	}
}
