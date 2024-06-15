package bodyParams

import (
	"errors"
	"net/http"
)

type RefreshParams struct {
	RefreshToken string `json:"refresh_token"`
}

func (p RefreshParams) Parse(r *http.Request) (RefreshParams, error) {
	params, err := Decode(r, RefreshParams{})
	if err != nil {
		return RefreshParams{}, err
	}
	if err := validateRefreshToken(params.RefreshToken); err != nil {
		return RefreshParams{}, err
	}
	return params, nil
}

func validateRefreshToken(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}
	return nil
}
