package bodyParams

import (
	"encoding/json"
	"net/http"
)

type BodyParser interface {
	Parse(r *http.Request) (BodyParser, error)
}

func Decode[T any](r *http.Request, paramsType T) (T, error) {
	decoder := json.NewDecoder(r.Body)
	var params T
	err := decoder.Decode(&params)
	if err != nil {
		return params, err
	}
	return params, nil
}
