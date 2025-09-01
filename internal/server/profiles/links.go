package profiles

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	errInvalidURL   = "URL inválida"
	errNameRequired = "El nombre es requerido"
)

type LinkRegisterInput struct {
	URL  string
	Name string
}

type LinkRegisterError struct {
	NameError string
	URLError  string
}

type LinkDeleteInput struct {
	LinkID string
}

type LinkRegisterData struct {
	URL  string
	Name string
	ID   string
}

func GetLinkRegisterInput(r *http.Request) (LinkRegisterInput, LinkRegisterError, bool) {
	errFound := false
	inputErrors := LinkRegisterError{}
	rawURL := r.FormValue("url")
	if _, err := url.Parse(rawURL); err != nil || rawURL == "" || len(rawURL) > 256 {
		errFound = true
		inputErrors.URLError = errInvalidURL
	}
	name := r.FormValue("name")
	if len(name) < 1 || len(name) > 256 {
		errFound = true
		inputErrors.NameError = errNameRequired
	}
	return LinkRegisterInput{
		URL:  rawURL,
		Name: name,
	}, inputErrors, errFound
}

func GetLinkDeleteInput(r *http.Request) (LinkDeleteInput, error) {
	linkID := chi.URLParam(r, "linkID")
	if err := shared.ValidateUUID(linkID); err != nil {
		return LinkDeleteInput{}, err
	}
	return LinkDeleteInput{
		LinkID: linkID,
	}, nil
}

type LinkStorage interface {
	RegisterLink(ctx context.Context, linkID, userID, url, name string) error
	DeleteLink(ctx context.Context, userID, linkID string) error
}

type registerLinkInputFn func(r *http.Request) (LinkRegisterInput, LinkRegisterError, bool)

func CreateRegisterLinkHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage LinkStorage, inputFn registerLinkInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, inputErr, errFound := inputFn(r)
		if errFound {
			if err := templ.Render(w, "linkErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		linkID := uuid.New().String()
		if err := storage.RegisterLink(r.Context(), linkID, user.ID, input.URL, input.Name); err != nil {
			inputErr.NameError = errUnexpected
			if err := templ.Render(w, "linkErrors", inputErr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		data := LinkRegisterData{
			URL:  input.URL,
			Name: input.Name,
			ID:   linkID,
		}

		w.Header().Set("HX-Trigger", "link-added")
		if err := templ.Render(w, "linkEntry", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type deleteLinkInputFn func(r *http.Request) (LinkDeleteInput, error)

func CreateDeleteLinkHandler(auth shared.AuthRep, storage LinkStorage, inputFn deleteLinkInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := storage.DeleteLink(r.Context(), user.ID, input.LinkID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
