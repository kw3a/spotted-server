package profiles

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errUserTaken         = "Nombre de usuario no disponible"
	errNameLength        = "Debe tener entre 3 a 255 caracteres"
	errDescriptionLength = "Debe tener entre 20 a 500 caracteres"
)

type UserStorage interface {
	CreateUser(ctx context.Context, id, nick, name, password string) error
	Save(ctx context.Context, refreshToken string) error
}

type UserInput struct {
	Name     string
	Password string
	Nick     string
}

type UserInputErrors struct {
	NameError     string
	PasswordError string
	NickError     string
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
	nick := r.FormValue("nick")
	if len(nick) < 3 || len(nick) > 32 {
		inputErrors.NickError = shared.ErrLength(3, 32)
		inputErrFound = true
	}

	return UserInput{
		Name:     name,
		Password: password,
		Nick:     nick,
	}, inputErrors, inputErrFound
}

type userInputFunc func(*http.Request) (UserInput, UserInputErrors, bool)

func CreateUserHandler(
	authType LoginAuthType,
	templ shared.TemplatesRepo,
	storage UserStorage,
	inputFn userInputFunc,
	redirection string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, inputErr, errorExists := inputFn(r)
		if errorExists {
			renderErr := templ.Render(w, "userFormErrors", inputErr)
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		userID := uuid.NewString()
		err := storage.CreateUser(r.Context(), userID, input.Nick, input.Name, input.Password)
		if err != nil {
			if strings.Contains(err.Error(), "1062") {
				renderErr := templ.Render(w, "userFormErrors", UserInputErrors{NickError: errUserTaken})
				if renderErr != nil {
					http.Error(w, renderErr.Error(), http.StatusInternalServerError)
				}
				return
			}
			renderErr := templ.Render(w, "userFormErrors", UserInputErrors{NickError: errUnexpected + err.Error()})
			if renderErr != nil {
				http.Error(w, renderErr.Error(), http.StatusInternalServerError)
			}
			return
		}
		refreshToken, accessToken, err := authType.CreateTokens(userID)
		if err != nil {
			inputErr.NickError = errUnexpected
			if err := templ.Render(w, "userFormErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.Save(r.Context(), refreshToken)
		if err != nil {
			inputErr.NickError = errUnexpected
			if err := templ.Render(w, "userFormErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		auth.SetTokenCookie(w, "refresh_token", refreshToken)
		auth.SetTokenCookie(w, "access_token", accessToken)
		redirRoute := redirection + userID
		w.Header().Set("HX-Redirect", redirRoute)
		w.WriteHeader(http.StatusCreated)
	}
}
