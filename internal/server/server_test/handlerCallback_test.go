package servertest

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
)

func TestCallbackUrlParams(t *testing.T) {
	submissionID := uuid.NewString()
	tcID := uuid.NewString()
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": submissionID,
		"testCaseID":   tcID,
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	params, err := server.NewCallbackUrlParams(reqWithUrlParam)
	if err != nil {
		t.Error(err)
	}
	if params.SubmissionID != submissionID {
		t.Error("invalid submission ID")
	}
	if params.TestCaseID != tcID {
		t.Error("invalid tc ID")
	}
}

func TestCallbackInputValid(t *testing.T) {
	input := server.CallbackInput{}
	problems := input.Valid(context.Background())
	if len(problems) != 0 {
		t.Error(problems)
	}
}
