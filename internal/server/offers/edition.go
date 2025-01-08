package offers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type OfferEdition struct {
	OfferID     string
	Languages   []int32
	Duration    int32
}

func GetOfferEditionInput(r *http.Request) (OfferEdition, error) {
	err := r.ParseForm()
	if err != nil {
		return OfferEdition{}, err
	}
	offerID := chi.URLParam(r, "offerID")
	if err := shared.ValidateUUID(offerID); err != nil {
		return OfferEdition{}, err
	}
	languages := r.Form["languages"]
	if len(languages) == 0 {
		return OfferEdition{}, fmt.Errorf("languages are empty")
	}
	fmt.Println(languages)
	intLanguages := []int32{}
	for _, lang := range languages {
		intLang, err := strconv.Atoi(lang)
		if err != nil {
			return OfferEdition{}, err
		}
		intLanguages = append(intLanguages, int32(intLang))
	}
	duration, err := strconv.Atoi(r.FormValue("duration"))
	if err != nil {
		return OfferEdition{}, err
	}
	return OfferEdition{
		OfferID:     offerID,
		Languages:   intLanguages,
		Duration:    int32(duration),
	}, nil
}

type OfferEditionStorage interface {
	InsertQuiz(r *http.Request, quizID, offerID string, languages []int32, duration int32) error
}

type offerEditionInputFn func(r *http.Request) (OfferEdition, error)
func CreateOfferEdition(
	storage OfferEditionStorage,
	templ shared.TemplatesRepo,
	inputFn offerEditionInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quizID := uuid.New().String()
		err = storage.InsertQuiz(r, quizID, input.OfferID, input.Languages, input.Duration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
