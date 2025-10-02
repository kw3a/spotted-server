package companies

import (
	"context"
	"net/http"
	"net/url"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errFormSize     = "La solicitud excede el tamaño máximo de 10 MB"
	errURLFormat    = "URL inválida"
	errImageFormat  = "Formato de imagen inválido"
	errImageMissing = "Debe seleccionar una imagen"
	errImageSave    = "La imagen no se pudo guardar"
)

type CompanyRegInput struct {
	Name        string
	Description string
	Website     string
	ImageURL    string
}

type CompanyRegErrors struct {
	NameError        string
	DescriptionError string
	WebsiteError     string
	ImageURLError    string
}

func GetRegisterCompanyInput(
	cloudinaryService shared.CloudinaryService,
	r *http.Request,
) (CompanyRegInput, CompanyRegErrors, bool) {
	errFound := false
	inputErrors := CompanyRegErrors{}
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		inputErrors.ImageURLError = errFormSize
		return CompanyRegInput{}, inputErrors, true
	}
	name := r.FormValue("name")
	if len(name) < 3 || len(name) > 64 {
		inputErrors.NameError = shared.ErrLength(3, 64)
		errFound = true
	}
	description := r.FormValue("description")
	if len(description) < 16 || len(description) > 500 {
		inputErrors.DescriptionError = shared.ErrLength(16, 500)
		errFound = true
	}
	website := r.FormValue("website")
	if website != "" {
		u, err := url.Parse(website)
		if err != nil || u.Scheme == "" || u.Host == "" {
			inputErrors.WebsiteError = errURLFormat
			errFound = true
		} else if len(website) > 256 {
			inputErrors.WebsiteError = shared.ErrLength(9, 256)
			errFound = true
		}
	}
	image, _, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			inputErrors.ImageURLError = errImageMissing
			errFound = true
		} else {
			inputErrors.ImageURLError = errImageFormat
			errFound = true
		}
	}
	var imageURL string
	if err == nil && !errFound {
		defer image.Close()
		resp, err := cloudinaryService.Upload(r.Context(), image, uploader.UploadParams{})
		if err != nil {
			inputErrors.ImageURLError = errImageSave
			errFound = true
		}
		imageURL = resp.SecureURL
	}
	return CompanyRegInput{
		Name:        name,
		Description: description,
		Website:     website,
		ImageURL:    imageURL,
	}, inputErrors, errFound
}

type CompanyStorage interface {
	RegisterCompany(ctx context.Context, id, userID, name, description, website, imageURL string) error
}

type registerCompanyInputFn func(shared.CloudinaryService, *http.Request) (CompanyRegInput, CompanyRegErrors, bool)

func CreateRegisterHandler(
	storage CompanyStorage,
	auth shared.AuthRep,
	cloudinaryService shared.CloudinaryService,
	inputFn registerCompanyInputFn,
	templ shared.TemplatesRepo,
	redirectPath string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errorExists := inputFn(cloudinaryService, r)
		if errorExists {
			if err := templ.Render(w, "companyRegErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		companyID := uuid.New().String()
		err = storage.RegisterCompany(r.Context(), companyID, user.ID, input.Name, input.Description, input.Website, input.ImageURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("HX-Redirect", redirectPath+companyID)
		w.WriteHeader(http.StatusOK)
	}
}
