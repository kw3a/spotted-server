package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ResultsInput struct {
	SubmissionID string
}

func GetResultsInput(r *http.Request) (ResultsInput, error) {
	submissionID := chi.URLParam(r, "submissionID")
	if err := ValidateUUID(submissionID); err != nil {
		return ResultsInput{}, fmt.Errorf("invalid submission ID: %w", err)
	}
	input := ResultsInput{
		SubmissionID: submissionID,
	}
	return input, nil
}

type resultsInputFn func(r *http.Request) (ResultsInput, error)
func CreateJudgeResultsHandler(st StreamService, inputFn resultsInputFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := inputFn(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		listener, err := st.Listen(input.SubmissionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		SSEHeaders(w)
		EventStream(w, listener, FormatSSEvent)
	}
}

type formatFunc = func(string, string) string

func EventStream(w http.ResponseWriter, listenerChannel chan string, formatter formatFunc) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "sse is not suppported", http.StatusInternalServerError)
		return
	}
	last := ""
	for msg := range listenerChannel {
		last = msg
		formatedEvent := formatter("result", msg)
		_, err := fmt.Fprint(w, formatedEvent)
		if err != nil {
			log.Println(err)
			break
		}
		flusher.Flush()
	}
	formatedEvent := formatter("finished", last)
	_, err := fmt.Fprint(w, formatedEvent)
	if err != nil {
		log.Println(err)
	}
	flusher.Flush()
}

func FormatSSEvent(event string, data string) string {
	formatData := fmt.Sprintf("<p>%s</p>", data)
	if event == "finished" {
		formatData += `<div hx-on::load="htmx.trigger('#score', 'evtrunfinished')"></div>`
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, formatData)
}

func SSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func (app *App) ResultsHandler() http.HandlerFunc {
	return CreateJudgeResultsHandler(
		app.Stream,
		GetResultsInput,
	)
}
