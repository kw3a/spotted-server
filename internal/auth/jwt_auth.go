package auth

type JWTAuth struct {
	secret string
}

func NewJWTAuth(jwtSecret string) *JWTAuth {
	return &JWTAuth{secret: jwtSecret}
}

func (jwtAuth *JWTAuth) Authenticate(accessToken string) (string, error) {
	parsedToken, err := ValidParsedToken(accessToken, jwtAuth.secret, tokenTypeAccess)
	if err != nil {
		return "", err
	}
	return parsedToken.userID, nil
}
func (jwtAuth *JWTAuth) CreateAccess(refreshToken string) (string, error) {
	parsedToken, err := ValidParsedToken(refreshToken, jwtAuth.secret, tokenTypeRefresh)
	if err != nil {
		return "", err
	}
	return newJWT(parsedToken.userID, jwtAuth.secret, tokenTypeAccess)
}
func (jwtAuth *JWTAuth) CreateRefresh(userID string) (string, error) {
	err := validateUserID(userID)
	if err != nil {
		return "", err
	}
	return newJWT(userID, jwtAuth.secret, tokenTypeRefresh)
}
func (jwtAuth *JWTAuth) ValidateRefresh(refreshToken string) error {
	_, err := ValidParsedToken(refreshToken, jwtAuth.secret, tokenTypeRefresh)
	return err
}
