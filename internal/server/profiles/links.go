package profiles

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type LinkRegisterInput struct {
	URL  string
	Name string
}

type LinkDeleteInput struct {
	LinkID string
}

type LinkRegisterData struct {
	URL  string
	Name string
	ID   string
}

func GetLinkRegisterInput(r *http.Request) (LinkRegisterInput, error) {
	rawURL := r.FormValue("url")
	if _, err := url.Parse(rawURL); err != nil {
		return LinkRegisterInput{}, err
	}
	name := r.FormValue("name")
	if name == "" {
		return LinkRegisterInput{}, fmt.Errorf("name is required")
	}
	return LinkRegisterInput{
		URL:  rawURL,
		Name: name,
	}, nil
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

type registerLinkInputFn func(r *http.Request) (LinkRegisterInput, error)

func CreateRegisterLinkHandler(templ shared.TemplatesRepo, auth shared.AuthRep, storage LinkStorage, inputFn registerLinkInputFn) http.HandlerFunc {
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
		linkID := uuid.New().String()
		err = storage.RegisterLink(r.Context(), linkID, user.ID, input.URL, input.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := LinkRegisterData{
			URL:  input.URL,
			Name: input.Name,
			ID:   linkID,
		}
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
		err = storage.DeleteLink(r.Context(), user.ID, input.LinkID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

