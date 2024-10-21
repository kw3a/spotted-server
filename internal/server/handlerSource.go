package server

import (
	"context"
	"net/http"
)

type SourceStorage interface {
	LastSrc(ctx context.Context, userID, problemID string, languageID int32) (string, error)
}

type SourceInput struct {
	ProblemID  string
	LanguageID int32
}

func GetSourceInput(r *http.Request) (SourceInput, error) {
	problemID := r.FormValue("problemID")
	if err := ValidateUUID(problemID); err != nil {
		return SourceInput{}, err
	}
	languageID := r.FormValue("languageID")
	languageIDInt32, err := ValidateLanguageID(languageID)
	if err != nil {
		return SourceInput{}, err
	}
	return SourceInput{
		ProblemID:  problemID,
		LanguageID: languageIDInt32,
	}, nil
}

type sourceInputFunc func(r *http.Request) (SourceInput, error)

func CreateSourceHandler(storage SourceStorage, authServ AuthRep, inputFn sourceInputFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authServ.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		src, err := storage.LastSrc(r.Context(), user.ID, input.ProblemID, input.LanguageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, err = w.Write([]byte(src))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (DI *App) SourceHandler() http.HandlerFunc {
	return CreateSourceHandler(
		DI.Storage,
		DI.AuthService,
		GetSourceInput,
	)
}
