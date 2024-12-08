package companies

import (
	"context"
	"net/http"
	"net/url"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type RegisterCompanyInput struct {
	Name        string
	Description string
	Website     string
	ImageURL    string
}

func GetRegisterCompanyInput(cloudinaryService shared.CloudinaryService, r *http.Request) (RegisterCompanyInput, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		return RegisterCompanyInput{}, err
	}
	name := r.FormValue("name")
	if name == "" {
		return RegisterCompanyInput{}, err
	}
	description := r.FormValue("description")
	if description == "" {
		return RegisterCompanyInput{}, err
	}
	website := r.FormValue("website")
	_, err = url.Parse(website)
	if err != nil {
		return RegisterCompanyInput{}, err
	}
	image, _, err := r.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return RegisterCompanyInput{}, err
	}
	var imageURL string
	if err == nil {
		defer image.Close()
		resp, err := cloudinaryService.Upload(r.Context(), image, uploader.UploadParams{})
		if err != nil {
			return RegisterCompanyInput{}, err
		}
		imageURL = resp.SecureURL
	}
	return RegisterCompanyInput{
		Name:        name,
		Description: description,
		Website:     website,
		ImageURL:    imageURL,
	}, nil
}

type CompanyStorage interface {
	RegisterCompany(ctx context.Context, id, userID, name, description, website, imageURL string) error
}

type registerCompanyInputFn func(shared.CloudinaryService, *http.Request) (RegisterCompanyInput, error)

func CreateRegisterCompanyHandler(storage CompanyStorage, auth shared.AuthRep, cloudinaryService shared.CloudinaryService, inputFn registerCompanyInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(cloudinaryService, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		companyID := uuid.New().String()
		err = storage.RegisterCompany(r.Context(), companyID, user.ID, input.Name, input.Description, input.Website, input.ImageURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("HX-Redirect", "/companies/"+companyID)
		w.WriteHeader(http.StatusOK)
	}
}
