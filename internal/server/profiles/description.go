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

type DescrStorage interface {
	UpdateDescription(ctx context.Context, userID, description string) error
}

func GetDescUpdateInput(r *http.Request) (DescUpdateInput, DescUpdateErrors, bool) {
	inputErrors := DescUpdateErrors{}
	errFound := false
	res := DescUpdateInput{}

	description := r.FormValue("description")
	if len(description) > 128 || len(description) < 5 {
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
			if err := templ.Render(w, "descrAlerts", shared.Alert{Ok: false, Msg: inputErr.DescrError}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err := storage.UpdateDescription(r.Context(), user.ID, input.Description); err != nil {
			if err := templ.Render(w, "descrAlerts", shared.Alert{Ok: false, Msg: errUnexpected}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err := templ.Render(w, "descrAlerts", shared.Alert{Ok: true, Msg: shared.MsgSaved}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
