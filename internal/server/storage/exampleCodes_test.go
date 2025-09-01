package storage

import (
	"testing"
)

func TestExampleCodeWrongLanguageID(t *testing.T) {
	_, err := ExampleCode(0)
	if err == nil {
		t.Error("expected error")
	}
}

func TestExampleCodeGo(t *testing.T) {
	code, err := ExampleCode(60)
	if err != nil {
		t.Error(err)
	}
	if code == "" {
		t.Error("empty code")
	}
}
