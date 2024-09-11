package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type AuthRep interface {
	GetUser(r *http.Request) (userID string, err error)
}
func ValidateUUID(id string) error {
	return uuid.Validate(id)
}
type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

type ErrorMesage struct {
	Msg string `json:"msg"`
}

func ValidateLanguageID(languageID string) (int32, error) {
	languageIDInt, err := strconv.ParseInt(languageID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("languageID is not a valid integer")
	}
	languageIDInt32 := int32(languageIDInt)
	if languageIDInt32 < 0 || languageIDInt32 > 100 {
		return 0, fmt.Errorf("languageID is not in the valid range")
	}
	return languageIDInt32, nil
}

func DecodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
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
