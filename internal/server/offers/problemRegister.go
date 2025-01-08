package offers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

type ProblemRegInput struct {
	QuizID      string
	Title       string
	Description string
	TimeLimit   int32
}

type ProblemRegErr struct {
	TitleError       string
	DescriptionError string
	TimeLimitError   string
}

type ProblemRegData struct {
	ProblemID string
}

type ProblemRegStorage interface {
	InsertProblem(ctx context.Context, problemID, quizID, title, description string, timeLimit, memoryLimit int32) error
}

type problemRegInputFn func(r *http.Request) (ProblemRegInput, ProblemRegErr, bool)

func CreateProblemRegistrationHandler(
	templ shared.TemplatesRepo,
	storage ProblemRegStorage,
	inputFn problemRegInputFn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, inputErrors, errorExists := inputFn(r)
		if errorExists {
			if err := templ.Render(w, "problemFormErrors", inputErrors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		problemID := uuid.New().String()
		err := storage.InsertProblem(r.Context(), problemID, input.QuizID, input.Title, input.Description, input.TimeLimit, 262144)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := ProblemRegData{ProblemID: problemID}
		if err := templ.Render(w, "testCaseForm", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
