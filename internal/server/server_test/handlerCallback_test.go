package servertest

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server"
)



type Params map[string]string

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

func WithUrlParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return req
}

func WithUrlParams(r *http.Request, params Params) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}
	return req
}
