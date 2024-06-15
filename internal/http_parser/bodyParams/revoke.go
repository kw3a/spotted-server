package bodyParams

import "net/http"

type RevokeParams struct {
	RefreshToken string `json:"refresh_token"`
}

func (p RevokeParams) Parse(r *http.Request) (RevokeParams, error) {
	params, err := Decode(r, RevokeParams{})
	if err != nil {
		return RevokeParams{}, err
	}
	if err := validateRefreshToken(params.RefreshToken); err != nil {
		return RevokeParams{}, err
	}
	return params, nil
}
