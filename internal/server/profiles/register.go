package profiles

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errUserTaken         = "El correo ya está siendo utilizado"
	errNameLength        = "Debe tener entre 3 a 255 caracteres"
	errDescriptionLength = "Debe tener entre 20 a 500 caracteres"
)

type UserStorage interface {
	CreateUser(ctx context.Context, name, password, email, description string) error
}

type UserInput struct {
	Name        string
	Password    string
	Email       string
	Description string
}

type UserInputErrors struct {
	NameError        string
	PasswordError    string
	EmailError       string
	DescriptionError string
}

type CloudinaryService interface {
	Upload(
		ctx context.Context,
		file interface{},
		uploadParams uploader.UploadParams,
	) (*uploader.UploadResult, error)
}

func GetUserInput(r *http.Request) (UserInput, UserInputErrors, bool) {
	inputErrors := UserInputErrors{}
	inputErrFound := false
	name := r.FormValue("name")
	if len(name) < 3 || len(name) > 255 {
		inputErrors.NameError = shared.ErrLength(3, 255)
		inputErrFound = true
	}
	password := r.FormValue("password")
	if len(password) < 5 || len(password) > 30 {
		inputErrors.PasswordError = shared.ErrLength(5, 30)
		inputErrFound = true
	}
	email := r.FormValue("email")
	if strErr := EmailValidation(email); strErr != "" {
		inputErrors.EmailError = strErr
		inputErrFound = true
	}
	description := r.FormValue("description")
	if len(description) < 200 || len(description) > 500 {
		inputErrors.DescriptionError = shared.ErrLength(200, 500)
		inputErrFound = true
	}

	return UserInput{
		Name:        name,
		Password:    password,
		Email:       email,
		Description: description,
	}, inputErrors, inputErrFound
}

type userInputFunc func(*http.Request) (UserInput, UserInputErrors, bool)

func CreateUserHandler(templ shared.TemplatesRepo, storage UserStorage, inputFn userInputFunc, redirection string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, inputErr, errorExists := inputFn(r)
		if errorExists {
			renderErr := templ.Render(w, "userFormErrors", inputErr)
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		err := storage.CreateUser(r.Context(), input.Name, input.Password, input.Email, input.Description)
		if err != nil {
			if strings.Contains(err.Error(), "1062") {
				renderErr := templ.Render(w, "userFormErrors", UserInputErrors{EmailError: errUserTaken})
				if renderErr != nil {
					http.Error(w, renderErr.Error(), http.StatusInternalServerError)
				}
				return
			}
			renderErr := templ.Render(w, "userFormErrors", UserInputErrors{EmailError: errUnexpected + err.Error()})
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("HX-Redirect", redirection)
		w.WriteHeader(http.StatusCreated)
	}
}
