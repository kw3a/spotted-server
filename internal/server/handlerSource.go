package server

import (
	"context"
	"net/http"
	"strconv"
)

type SourceStorage interface {
	LastSrc(ctx context.Context, userID, problemID string, languageID int32) (string, error)
}

type AuthRep interface {
	GetUser(r *http.Request) (userID string, err error)
}

func createSourceHandler(storage SourceStorage, authServ AuthRep) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authServ.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		problemID := r.FormValue("problemID")
		if problemID == "" {
			http.Error(w, "invalid problemID", http.StatusBadRequest)
			return
		}
		languageID := r.FormValue("languageID")
		if languageID == "" {
			http.Error(w, "invalid languageID", http.StatusBadRequest)
			return
		}
		languageIDInt, err := strconv.Atoi(languageID)
		if err != nil {
			http.Error(w, "invalid languageID", http.StatusBadRequest)
			return
		}
		src, err := storage.LastSrc(r.Context(), userID, problemID, int32(languageIDInt))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(src))
	}
}

func (DI *App) SourceHandler() http.HandlerFunc {
	return createSourceHandler(
		DI.Storage,
		DI.AuthService,
	)
}
