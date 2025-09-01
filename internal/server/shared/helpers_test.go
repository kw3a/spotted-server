package shared

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateUUIDinvalidUUID(t *testing.T) {
	err := ValidateUUID("")
	if err == nil {
		t.Error("expected error, got nil")
	}
	err = ValidateUUID("invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestValidateUUID(t *testing.T) {
	id := uuid.NewString()
	err := ValidateUUID(id)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidaLanguageIDinvalidLanguageID(t *testing.T) {
	_, err := ValidateLanguageID("")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = ValidateLanguageID("invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = ValidateLanguageID("-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_, err = ValidateLanguageID("101")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestValidateLanguageID(t *testing.T) {
	languageID := "60"
	languageIDInt32, err := ValidateLanguageID(languageID)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if languageIDInt32 != 60 {
		t.Errorf("expected %d, got %d", 60, languageIDInt32)
	}
}
