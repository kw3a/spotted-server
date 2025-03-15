package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/quizes"
)

func (DI *App) SourceHandler() http.HandlerFunc {
	return quizes.CreateSrcHandler(
		quizes.GetSrcInput,
		DI.Storage,
		DI.Templ,
	)
}
