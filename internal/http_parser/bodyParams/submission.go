package bodyParams

import (
	"encoding/base64"
	"net/http"
)

type SubmissionParams struct {
	Src        string `json:"src"`
	LanguageID int32  `json:"language_id"`
}

func (s SubmissionParams) Parse(r *http.Request) (SubmissionParams, error) {
	params, err := Decode(r, SubmissionParams{})
	if err != nil {
		return SubmissionParams{}, err
	}
	_, err = base64.StdEncoding.DecodeString(params.Src)
	if err != nil {
		return SubmissionParams{}, err
	}
	return params, nil
}
