package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	errUserID         = errors.New("user ID is not valid")
	errTokenType      = errors.New("token type is not valid")
	errExpirationTime = errors.New("the token is expired")
)

type tokenType string

const (
	tokenTypeAccess  tokenType = "access"
	tokenTypeRefresh tokenType = "refresh"
)

type ParsedToken struct {
	userID         string
	tokenType      tokenType
	expirationTime time.Time
}

func parseToken(jwtStr, secret string) (*ParsedToken, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		jwtStr,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil },
	)
	if err != nil {
		return &ParsedToken{}, err
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return &ParsedToken{}, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return &ParsedToken{}, err
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return &ParsedToken{}, err
	}
	expTime := exp.Time

	return &ParsedToken{
		userID:         userID,
		tokenType:      tokenType(issuer),
		expirationTime: expTime,
	}, nil
}

func ValidParsedToken(jwtStr, secret string, tokenType tokenType) (*ParsedToken, error) {
	parsedToken, err := parseToken(jwtStr, secret)
	if err != nil {
		return &ParsedToken{}, err
	}
	if err := parsedToken.IsValid(tokenType); err != nil {
		return &ParsedToken{}, err
	}
	return parsedToken, nil
}

func (t *ParsedToken) IsValid(objective tokenType) error {
	if err := t.isCorrectType(objective); err != nil {
		return err
	}
	if err := t.isExpired(); err != nil {
		return err
	}
	if err := t.isUserValid(); err != nil {
		return err
	}
	return nil
}

func (t *ParsedToken) isCorrectType(objective tokenType) error {
	if t.tokenType != objective {
		return errTokenType
	}
	return nil
}

func (t *ParsedToken) isExpired() error {
	if time.Now().UTC().After(t.expirationTime) {
		return errExpirationTime
	}
	return nil
}

func (t *ParsedToken) isUserValid() error {
	return validateUserID(t.userID)
}

func validateUserID(userID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		return errUserID
	}
	return nil
}
