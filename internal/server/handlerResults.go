package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

func createJudgeResultsHandler(st *codejudge.Stream) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		submissionID := chi.URLParam(r, "submissionID")
		if uuid.Validate(submissionID) != nil {
      http.Error(w, "invalid submissionID", http.StatusBadRequest)
			return
		}
		topic, err := st.GetTopic(submissionID)
		if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		EventStream(w, topic.Listen())
	}
}

func EventStream(w http.ResponseWriter, listenerChannel chan string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
    http.Error(w, "sse is not suppported", http.StatusInternalServerError)
		return
	}
  last := ""
	for msg := range listenerChannel {
    last = msg
		err := SSEWriter(w, "result", msg)
		if err != nil {
			log.Println(err)
			break
		}
		flusher.Flush()
	}
	err := SSEWriter(w, "finished", last)
  if err != nil {
    log.Println(err)
  }
	flusher.Flush()
}

// SSEWriter calls FormatServerSentEvent and adds \n on the end of all the events (event1\n\n\nevent2\n\n\n...)
func SSEWriter(w http.ResponseWriter, event, data string) error {
	SSEHeaders(w)
	formatedEvent, err := FormatServerSentEvent(event, data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, formatedEvent)
	return err
}

func FormatServerSentEvent(event string, data string) (string, error) {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, data), nil
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
