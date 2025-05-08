package profiles

import (
	"context"
	"net/http"
	"strconv"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errInvalidCell = "número inválido"
)

type CellUpdateInput struct {
	Cell string
}

type CellUpdateErrors struct {
	CellError string
}

type CellUpdateStorage interface {
	UpdateCell(ctx context.Context, userID, cell string) error
}

func CellValidation(cell string) string {
	intCell, err := strconv.Atoi(cell)
	if err != nil || intCell < 0 || len(cell) < 6 || len(cell) > 14 {
		return errInvalidCell
	}
	return ""
}

func GetUpdateCellInput(r *http.Request) (CellUpdateInput, CellUpdateErrors, bool) {
	inputErrors := CellUpdateErrors{}
	errFound := false

	cell := r.FormValue("cell")
	if strErr := CellValidation(cell); strErr != "" {
		inputErrors.CellError = strErr
		errFound = true
	}
	return CellUpdateInput{
		Cell: cell,
	}, inputErrors, errFound
}

type cellUpdateInputFn func(r *http.Request) (CellUpdateInput, CellUpdateErrors, bool)

func CreateUpdateCellHandler(
	templ shared.TemplatesRepo,
	auth shared.AuthRep,
	storage CellUpdateStorage,
	inputFn cellUpdateInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "cellAlerts", shared.Alert{Ok: false, Msg: inputErr.CellError}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.UpdateCell(r.Context(), user.ID, input.Cell)
		if err != nil {
			if err := templ.Render(w, "cellAlerts", shared.Alert{Ok: false, Msg: errUnexpected}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err := templ.Render(w, "cellAlerts", shared.Alert{Ok: true, Msg: shared.MsgSaved}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
