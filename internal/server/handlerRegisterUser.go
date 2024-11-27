package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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

func NameValidation(name string) error {
	if len(name) < 3 || len(name) > 255 {
		return fmt.Errorf("name length must be less than 255 and more than 3")
	}
	return nil
}

func DescriptionValidation(description string) error {
	if len(description) < 20 || len(description) > 500 {
		return fmt.Errorf("description length must be less than 500 and more than 20")
	}
	return nil
}

type CloudinaryService interface {
	Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error)
}

func GetUserInput(r *http.Request) (UserInput, UserInputErrors, error) {
	name := r.FormValue("name")
	if err := NameValidation(name); err != nil {
		return UserInput{}, UserInputErrors{NameError: err.Error()}, err
	}
	password := r.FormValue("password")
	if err := PasswordValidation(password); err != nil {
		return UserInput{}, UserInputErrors{PasswordError: err.Error()}, err
	}
	email := r.FormValue("email")
	if err := EmailValidation(email); err != nil {
		return UserInput{}, UserInputErrors{EmailError: err.Error()}, err
	}
	description := r.FormValue("description")
	if err := DescriptionValidation(description); err != nil {
		return UserInput{}, UserInputErrors{DescriptionError: err.Error()}, err
	}

	return UserInput{
		Name:        name,
		Password:    password,
		Email:       email,
		Description: description,
	}, UserInputErrors{}, nil
}

type userInputFunc func(*http.Request) (UserInput, UserInputErrors, error)

func CreateUserHandler(templ TemplatesRepo, storage UserStorage, inputFn userInputFunc, redirection string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, inputErr, err := inputFn(r)
		if err != nil {
			renderErr := templ.Render(w, "userFormErrors", inputErr)
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.CreateUser(r.Context(), input.Name, input.Password, input.Email, input.Description)
		if err != nil {
			renderErr := templ.Render(w, "userFormErrors", UserInputErrors{EmailError: err.Error()})
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("HX-Redirect", redirection)
		w.WriteHeader(http.StatusCreated)
	}
}

func (DI *App) UserHandler() http.HandlerFunc {
	return CreateUserHandler(DI.Templ, DI.Storage, GetUserInput, "/login")
}
