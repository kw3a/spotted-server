package shared

import (
	"context"
	"io"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/kw3a/spotted-server/internal/auth"
)

type CloudinaryService interface {
	Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error)
}

type AuthRep interface {
	GetUser(r *http.Request) (userID auth.AuthUser, err error)
}

type TemplatesRepo interface {
	Render(w io.Writer, name string, data interface{}) error
}
