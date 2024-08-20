package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
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

func createJudgeResultsHandler(st *codejudge.Stream) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := GetResultsInput(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		topic, err := st.GetTopic(input.SubmissionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		SSEHeaders(w)
		EventStream(w, topic.Listen(), FormatSSEvent)
	}
}
type formatFunc = func(string, string) (string, error) 
func EventStream(w http.ResponseWriter, listenerChannel chan string, formatter formatFunc) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "sse is not suppported", http.StatusInternalServerError)
		return
	}
	last := ""
	for msg := range listenerChannel {
		last = msg
		formatedEvent, err := formatter("result", msg)
		if err != nil {
			log.Println(err)
			break
		}

		_, err = fmt.Fprint(w, formatedEvent)
		if err != nil {
			log.Println(err)
			break
		}
		flusher.Flush()
	}
	formatedEvent, err := formatter("finished", last)
	if err != nil {
		log.Println(err)
	}
	_, err = fmt.Fprint(w, formatedEvent)
	if err != nil {
		log.Println(err)
	}
	flusher.Flush()
}

func FormatSSEvent(event string, data string) (string, error) {
	if event == "" {
		return "", fmt.Errorf("event is empty")
	}
	if data == "" {
		return "", fmt.Errorf("data is empty")
	}
	formatData := fmt.Sprintf("<p>%s</p>", data)
	if event == "finished" {
		formatData += `<div hx-on::load="htmx.trigger('#score', 'evtrunfinished')"></div>`
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, formatData), nil
}

func SSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func (app *App) ResultsHandler() http.HandlerFunc {
	return createJudgeResultsHandler(
		app.Stream,
	)
}
