package offers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferListData struct {
	User     auth.AuthUser
	Offers   []shared.Offer
	Search   string
	NextPage int32
}

const (
	queryUpperLimit = 30
	errorQueryLimit = "el límite máximo es de 30 caracteres"
)

type OfferListStorage interface {
	SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error)
}

func GetListParams(r *http.Request) (shared.OfferQueryParams, error) {
	params := shared.OfferQueryParams{}
	q := r.URL.Query()
	query := q.Get("q")
	if len(query) > queryUpperLimit {
		return shared.OfferQueryParams{}, fmt.Errorf(errorQueryLimit)
	}
	if query != "" {
		params.Query = query
	}
	params.Page = shared.PageParam(r)
	return params, nil
}

type OfferListParamsFn func(r *http.Request) (shared.OfferQueryParams, error)

func CreateOfferListHandler(
	paramsFn OfferListParamsFn,
	authService shared.AuthRep,
	storage OfferListStorage,
	templ shared.TemplatesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authService.GetUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		params, err := paramsFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		offers, err := storage.SelectOffers(r.Context(), params)
		if err != nil {
			http.Error(w, "can't find offers", http.StatusInternalServerError)
			return
		}
		data := OfferListData{
			User:     user,
			Offers:   offers,
			NextPage: params.Page + 1,
		}
		if params.Query != "" {
			data.Search = params.Query
		}
		toRender := "offerListPage"
		if params.Page > 1 {
			toRender = "offerList"
		}
		if err = templ.Render(w, toRender, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
