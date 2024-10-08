package servertest

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
)

func TestCallbackUrlParamsEmpty(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	_, err = server.NewCallbackUrlParams(req)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallbackUrlParamsInvalidSubmissionID(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": "invalid",
		"testCaseID":   uuid.NewString(),
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	_, err = server.NewCallbackUrlParams(reqWithUrlParam)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCallbackUrlParamsInvalidTestCaseID(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Error(err)
	}
	urlParams := map[string]string{
		"submissionID": uuid.NewString(),
		"testCaseID":   "invalid",
	}
	reqWithUrlParam := WithUrlParams(req, urlParams)
	_, err = server.NewCallbackUrlParams(reqWithUrlParam)
	if err == nil {
		t.Error("expected error")
	}
}

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

func TestCallbackInputValidEmpty(t *testing.T) {
	input := server.CallbackInput{}
	problems := input.Valid(context.Background())
	if len(problems) != 0 {
		t.Error(problems)
	}
}

