package quizestest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/stretchr/testify/mock"
)

type authRepo struct{}

func (a authRepo) GetUser(r *http.Request) (userID auth.AuthUser, err error) {
	return auth.AuthUser{}, nil
}

type invalidAuthRepo struct{}

func (i invalidAuthRepo) GetUser(r *http.Request) (userID auth.AuthUser, err error) {
	return auth.AuthUser{}, errors.New("error")
}

type templates struct{}

func (t *templates) Render(w io.Writer, name string, data interface{}) error {
	return nil
}

type invalidTemplates struct{}

func (i invalidTemplates) Render(w io.Writer, name string, data interface{}) error {
	return errors.New("error")
}

type streamService struct {
	mock.Mock
}

func (s *streamService) Register(name string, tokens []string, duration time.Duration) error {
	args := s.Called(name, tokens, duration)
	return args.Error(0)
}

func (s *streamService) Listen(name string) (chan string, error) {
	args := s.Called(name)
	return args.Get(0).(chan string), args.Error(1)
}

func (s *streamService) Update(name, token, status string) error {
	args := s.Called(name, token, status)
	return args.Error(0)
}

type Params map[string]string

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

func formRequest(method, url string, formValues map[string][]string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil
	}
	req.Form = formValues
	return req
}

