package responseparser

import (
	"net/http"
)

type LoginResponseBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponseBody struct {
	AccessToken string `json:"access_token"`
}

func GetLoginResponseBody(accessToken string, refreshToken string) LoginResponseBody {
	return LoginResponseBody{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func GetRefreshResponseBody(accessToken string) RefreshResponseBody {
	return RefreshResponseBody{
		AccessToken: accessToken,
	}
}

func (res *LoginResponseBody) Send(w http.ResponseWriter) {
	RespondWithJSON(w, http.StatusOK, res)
}

func (res *RefreshResponseBody) Send(w http.ResponseWriter) {
	RespondWithJSON(w, http.StatusOK, res)
}
