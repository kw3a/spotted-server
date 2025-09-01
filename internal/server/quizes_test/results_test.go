package quizestest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/quizes"
)

func resultsInputFn(r *http.Request) (quizes.ResultsInput, error) {
	return quizes.ResultsInput{}, nil
}

func TestResultInputEmptySubmissionID(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	_, err := quizes.GetResultsInput(r)
	if err == nil {
		t.Error(err)
	}
}

func TestResultInputInvalidSubmissionID(t *testing.T) {
	params := map[string]string{
		"submissionID": "123",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	reqWithParams := WithUrlParams(req, params)
	_, err := quizes.GetResultsInput(reqWithParams)
	if err == nil {
		t.Error(err)
	}
}

func TestResultInput(t *testing.T) {
	id := uuid.NewString()
	params := map[string]string{
		"submissionID": id,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	reqWithParams := WithUrlParams(req, params)
	input, err := quizes.GetResultsInput(reqWithParams)
	if err != nil {
		t.Error(err)
	}
	if input.SubmissionID != id {
		t.Errorf("expected: %v \nobtained: %v", id, input.SubmissionID)
	}
}

func TestResultsHandlerBadInput(t *testing.T) {
	invalidInputFn := func(r *http.Request) (quizes.ResultsInput, error) {
		return quizes.ResultsInput{}, fmt.Errorf("error")
	}
	handler := quizes.CreateJudgeResultsHandler(&streamService{}, invalidInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("got: %v \nwant: %v", status, http.StatusBadRequest)
	}
}

func TestResultsHandlerBadStream(t *testing.T) {
	stream := new(streamService)
	stream.On("Listen", "").Return(make (chan string), fmt.Errorf("error"))
	handler := quizes.CreateJudgeResultsHandler(stream, resultsInputFn)
	if handler == nil {
		t.Error("handler is nil")
	}
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("got: %v \nwant: %v", status, http.StatusBadRequest)
	}
}

func TestSSEHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	quizes.SSEHeaders(rr)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("got: %v \nwant: %v", status, http.StatusOK)
	}
	expectedContentType := "text/event-stream"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("got: %v \nwant: %v", contentType, expectedContentType)
	}
	if conn := rr.Header().Get("Connection"); conn != "keep-alive" {
		t.Errorf("got: %v \nwant: %v", conn, "keep-alive")
	}
}

func TestFormatSSE(t *testing.T) {
	msg := "msg1"
	event1 := "result"
	event2 := "finished"
	formatedEvent := quizes.FormatSSEvent(event1, msg)
	expected := "event: result\ndata: <p>msg1</p>\n\n"
	if formatedEvent != expected {
		t.Errorf("expected: %v \nobtained: %v", expected, formatedEvent)
	}
	formatedEvent = quizes.FormatSSEvent(event2, msg)
	expected = "event: finished\ndata: <p>msg1</p><div hx-on::load=\"htmx.trigger('#score', 'evtrunfinished')\"></div>\n\n"
	if formatedEvent != expected {
		t.Errorf("expected: %v \nobtained: %v", expected, formatedEvent)
	}
}

func TestEventStream(t *testing.T) {
	msg := "msg1"
	msg2 := "msg2"
	testChannel := make(chan string)
	go func() {
		testChannel <- msg
		testChannel <- msg2
		close(testChannel)
	}()
	rr := httptest.NewRecorder()
	formatter := func(event string, data string) string {
		formated := event + ":" + data
		return formated
	}
	quizes.EventStream(rr, testChannel, formatter)
	res := rr.Body.String()
	expected := "result:msg1result:msg2finished:msg2"
	if res != expected {
		t.Errorf("expected: \n%s\nobtained: \n%s", expected, res)
	}
}
