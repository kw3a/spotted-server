package companiestest

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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

type authMock struct {
	mock.Mock
}

func (s *authMock) GetUser(r *http.Request) (userID auth.AuthUser, err error) {
	args := s.Called(r)
	return args.Get(0).(auth.AuthUser), args.Error(1)
}

type cldMock struct {
	mock.Mock
}
func (c *cldMock) Upload(
	ctx context.Context,
	file interface{},
	uploadParams uploader.UploadParams,
) (*uploader.UploadResult, error) {
	args := c.Called(ctx, file, uploadParams)
	return args.Get(0).(*uploader.UploadResult), args.Error(1)
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
