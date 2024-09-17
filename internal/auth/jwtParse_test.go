package auth

import "testing"

func TestParseJWTBadSecret(t *testing.T) {
	_, err := parseJWT("token", "")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestParseJWTBadToken(t *testing.T) {
	_, err := parseJWT("bad token", "secret")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
