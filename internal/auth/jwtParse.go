package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	errTokenType      = errors.New("token type is not valid")
	errExpirationTime = errors.New("the token is expired")
)

type tokenType string

const (
	tokenTypeAccess  tokenType = "access"
	tokenTypeRefresh tokenType = "refresh"
)

type JWToken struct {
	userID         string
	tokenType      tokenType
	expirationTime time.Time
}

func parseJWT(jwtStr, secret string) (*JWToken, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		jwtStr,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil },
	)
	if err != nil {
		return &JWToken{}, err
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return &JWToken{}, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return &JWToken{}, err
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return &JWToken{}, err
	}
	expTime := exp.Time

	return &JWToken{
		userID:         userID,
		tokenType:      tokenType(issuer),
		expirationTime: expTime,
	}, nil
}

func ValidJWT(jwtStr, secret string, tokenType tokenType) (*JWToken, error) {
	parsedToken, err := parseJWT(jwtStr, secret)
	if err != nil {
		return &JWToken{}, err
	}
	if err := parsedToken.isValid(tokenType); err != nil {
		return &JWToken{}, err
	}
	return parsedToken, nil
}

func (t *JWToken) isValid(objective tokenType) error {
	if t.tokenType != objective {
		return errTokenType
	}
	if time.Now().UTC().After(t.expirationTime) {
		return errExpirationTime
	}
	if err := validateUserID(t.userID); err != nil {
		return err
	}
	return nil
}

func validateUserID(userID string) error {
	return uuid.Validate(userID)
}
