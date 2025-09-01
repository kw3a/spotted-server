package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	ErrUserTaken = "El correo ya está siendo utilizado"
	ErrNotMatch       = "El correo y la contraseña no coinciden"
	ErrNotRegistered  = "Este usuario no existe"
	MsgSaved = "Guardado"
)

type Alert struct {
	Ok  bool
	Msg string
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
