package servertest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
)

func TestValidateUUIDinvalidUUID(t *testing.T) {
	err := server.ValidateUUID("")
	if err == nil {
		t.Error("expected error, got nil")
	}
	err = server.ValidateUUID("invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestValidateUUID(t *testing.T) {
	id := uuid.NewString()
	err := server.ValidateUUID(id)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidaLanguageIDinvalidLanguageID(t *testing.T) {
	_, err := server.ValidateLanguageID("")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = server.ValidateLanguageID("invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = server.ValidateLanguageID("-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = server.ValidateLanguageID("101")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestValidateLanguageID(t *testing.T) {
	languageID := "60"
	languageIDInt32, err := server.ValidateLanguageID(languageID)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if languageIDInt32 != 60 {
		t.Errorf("expected %d, got %d", 60, languageIDInt32)
	}
}
