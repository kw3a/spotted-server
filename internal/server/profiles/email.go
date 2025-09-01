package profiles

import (
	"context"
	"net/http"
	"regexp"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

type EmailUpdateInput struct {
	Email string
}

type EmailUpdateErrors struct {
	EmailError string
}

type EmailUpdateStorage interface {
	UpdateEmail(ctx context.Context, userID, email string) error
}

func EmailValidation(email string) string {
	if len(email) < 3 || len(email) > 254 {
		return errEmailInvalid
	}
	exp := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	compiledRegExp := regexp.MustCompile(exp)
	if !compiledRegExp.MatchString(email) {
		return errEmailInvalid
	}
	return ""
}

func GetEmailUpdateInput(r *http.Request) (EmailUpdateInput, EmailUpdateErrors, bool) {
	inputErrors := EmailUpdateErrors{}
	errFound := false

	email := r.FormValue("email")
	if strErr := EmailValidation(email); strErr != "" {
		inputErrors.EmailError = strErr
		errFound = true
	}
	return EmailUpdateInput{
		Email: email,
	}, inputErrors, errFound
}

type emailUpdateInputFn func(r *http.Request) (EmailUpdateInput, EmailUpdateErrors, bool)

func CreateUpdateEmailHandler(
	templ shared.TemplatesRepo,
	auth shared.AuthRep,
	storage EmailUpdateStorage,
	inputFn emailUpdateInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "emailAlerts", shared.Alert{Ok: false, Msg: inputErr.EmailError}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err = storage.UpdateEmail(r.Context(), user.ID, input.Email)
		if err != nil {
			if err := templ.Render(w, "emailAlerts", shared.Alert{Ok: false, Msg: errUnexpected}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err := templ.Render(w, "emailAlerts", shared.Alert{Ok: true, Msg: shared.MsgSaved}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
