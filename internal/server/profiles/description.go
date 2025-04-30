package profiles

import (
	"context"
	"net/http"
	"strings"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type DescUpdateInput struct {
	Description string
}

type DescUpdateErrors struct {
	DescrError string
}

type DescUpdateData struct {
	Description string
}

type DescrStorage interface {
	UpdateDescription(ctx context.Context, userID, description string) error
}

func GetDescUpdateInput(r *http.Request) (DescUpdateInput, DescUpdateErrors, bool) {
	inputErrors := DescUpdateErrors{}
	errFound := false
	res := DescUpdateInput{}

	description := r.FormValue("description")
	if len(description) > 164 || len(description) < 1{
		inputErrors.DescrError = shared.ErrLength(5, 128)
		errFound = true
		res.Description = ""
	} else {
		description = strings.TrimSpace(description)
		res.Description = description
	}
	return res, inputErrors, errFound

}

type descUpdateInputFn func(r *http.Request) (DescUpdateInput, DescUpdateErrors, bool)

func CreateDescUpdateHandler(
	templ shared.TemplatesRepo,
	auth shared.AuthRep,
	storage DescrStorage,
	inputFn descUpdateInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "descrErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err := storage.UpdateDescription(r.Context(), user.ID, input.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := DescUpdateData{
			Description: input.Description,
		}
		if err := templ.Render(w, "descrEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
