package bodyParams

import (
	"errors"
	"net/http"
	"regexp"
)

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p LoginParams) Parse(r *http.Request) (LoginParams, error) {
	params, err := Decode(r, LoginParams{})
	if err != nil {
		return LoginParams{}, err
	}
	if err := validateEmail(params.Email); err != nil {
		return LoginParams{}, err
	}
	if err := validatePassword(params.Password); err != nil {
		return LoginParams{}, err
	}
	return params, nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	exp := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	compiledRegExp := regexp.MustCompile(exp)
	if !compiledRegExp.MatchString(email) {
		return errors.New("invalid email address")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}
