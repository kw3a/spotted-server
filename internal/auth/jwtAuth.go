package auth

type JWTAuth struct {
	secret string
}

func NewJWTAuth(jwtSecret string) *JWTAuth {
	return &JWTAuth{secret: jwtSecret}
}

func (j *JWTAuth) WhoAmI(accessToken string) (string, error) {
	parsedToken, err := ValidJWT(accessToken, j.secret, tokenTypeAccess)
	if err != nil {
		return "", err
	}
	return parsedToken.userID, nil
}

func (j *JWTAuth) CreateTokens(userID string) (string, string, error) {
	err := validateUserID(userID)
	if err != nil {
		return "", "", err
	}
	refresh, err := newJWT(userID, j.secret, tokenTypeRefresh)
	if err != nil {
		return "", "", err
	}
	access, err := newJWT(userID, j.secret, tokenTypeAccess)
	if err != nil {
		return "", "", err
	}
	return refresh, access, nil
}
func (j *JWTAuth) CreateAccess(refreshToken string) (string, error) {
	parsedToken, err := ValidJWT(refreshToken, j.secret, tokenTypeRefresh)
	if err != nil {
		return "", err
	}
	return newJWT(parsedToken.userID, j.secret, tokenTypeAccess)
}
func (j *JWTAuth) CreateRefresh(userID string) (string, error) {
	err := validateUserID(userID)
	if err != nil {
		return "", err
	}
	return newJWT(userID, j.secret, tokenTypeRefresh)
}
func (j *JWTAuth) ValidateRefresh(refreshToken string) error {
	_, err := ValidJWT(refreshToken, j.secret, tokenTypeRefresh)
	return err
}
