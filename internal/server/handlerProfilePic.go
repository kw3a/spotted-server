package server

import (
	"context"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type ProfilePicInput struct {
	ImageURL string
}

func GetProfilePicInput(r *http.Request, cloudinaryService CloudinaryService) (ProfilePicInput, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		return ProfilePicInput{}, err
	}
	image, _, err := r.FormFile("image")
	if err != nil {
		return ProfilePicInput{}, err
	}
	defer image.Close()
	resp, err := cloudinaryService.Upload(r.Context(), image, uploader.UploadParams{})
	if err != nil {
		return ProfilePicInput{}, err
	}
	return ProfilePicInput{
		ImageURL: resp.SecureURL,
	}, nil
}

type ProfilePicStorage interface {
	UpdateProfilePic(ctx context.Context, userID, imageURL string) error
}

func CreateProfilePicHandler(storage ProfilePicStorage, auth AuthRep, cloudinaryService CloudinaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := GetProfilePicInput(r, cloudinaryService)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = storage.UpdateProfilePic(r.Context(), user.ID, input.ImageURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (DI *App) ProfilePicHandler() http.HandlerFunc {
	return CreateProfilePicHandler(DI.Storage, DI.AuthService, &DI.Cld.Upload)
}
