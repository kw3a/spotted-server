package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newJWT(userID, secret string, tokenType tokenType) (string, error) {
	signingKey := []byte(secret)
	expiresIn, err := expirationTime(tokenType)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(tokenType),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID,
	})
	return token.SignedString(signingKey)
}

func expirationTime(tokenType tokenType) (time.Duration, error) {
	switch tokenType {
	case tokenTypeAccess:
		return time.Hour * 6, nil
	case tokenTypeRefresh:
		return time.Hour * 24 * 120, nil
	default:
		return 0, errTokenType
	}
}
