package servertest

import (
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
)

func TestExampleCodeWrongLanguageID(t *testing.T) {
	_, err := server.ExampleCode(0)
	if err == nil {
		t.Error("expected error")
	}
}

func TestExampleCodeGo(t *testing.T) {
	code, err := server.ExampleCode(60)
	if err != nil {
		t.Error(err)
	}
	if code == "" {
		t.Error("empty code")
	}
}
