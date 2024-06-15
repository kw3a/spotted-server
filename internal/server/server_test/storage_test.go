package servertest 

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server"
)



func TestTCEmpty(t *testing.T) {
	dbTestCases := []database.GetTestCasesRow{}
	_, err := server.ToTC(dbTestCases)
  if err == nil {
    t.Errorf("Expected not nil, got nil")
  }
}

func TestToTC(t *testing.T) {
	dbTestCases := []database.GetTestCasesRow{
		{
			ID:          uuid.NewString(),
			TimeLimit:   70,
			MemoryLimit: 1024,
			Input:       "cHJpbnQoIkhlbGxvIik=",
			Output:      "cHJpbnQoIkhlbGxvIik=",
		},
	}
	res, err := server.ToTC(dbTestCases)
  if err != nil {
    t.Errorf("Expected nil, got %v", err)
  }
  if res == nil {
    t.Errorf("Expected not nil, got nil")
  }
}

