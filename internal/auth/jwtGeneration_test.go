package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewJWTBadToken(t *testing.T) {
	_, err := newJWT("userid", "secret", "bad token")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestNewJWT(t *testing.T) {
	userID1 := uuid.NewString()
	userID2 := uuid.NewString()
	token1, err := newJWT(userID1, "secret", "access")
	if err != nil {
		t.Errorf("NewJWT() failed: %v", err)
	}
	token2, err := newJWT(userID2, "secret", "access")
	if err != nil {
		t.Errorf("NewJWT() failed: %v", err)
	}
	if len(token1) != len(token2) {
		t.Errorf("expected same length, got %v and %v", len(token1), len(token2))
	}
}

func TestExpirationTimeBadToken(t *testing.T) {
	_, err := expirationTime("bad token")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestExpirationTime(t *testing.T) {
	expiresAt, err := expirationTime("access")
	if err != nil {
		t.Errorf("ExpirationTime() failed: %v", err)
	}
	if expiresAt != time.Hour*6 {
		t.Errorf("expected 6 hours, got %v", expiresAt)
	}
	expiresAt, err = expirationTime("refresh")
	if err != nil {
		t.Errorf("ExpirationTime() failed: %v", err)
	}
	if expiresAt != time.Hour*24*120 {
		t.Errorf("expected 120 days, got %v", expiresAt)
	}
}
