package servertest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kw3a/spotted-server/internal/server"
)


func TestSSEHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	server.SSEHeaders(rr)
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

func TestFormatSSEventEmpty(t *testing.T) {
	msg := "msg1"
	_, err := server.FormatSSEvent("", msg)
	if err == nil {
		t.Error(err)
	}
}

func TestFormatSSEMsgEmpty(t *testing.T) {
	msg := ""
	event := "result"
	_, err := server.FormatSSEvent(event, msg)
	if err == nil {
		t.Error(err)
	}
}

func TestFormatSSE(t *testing.T) {
	msg := "msg1"
	event1 := "result"
	event2 := "finished"
	formatedEvent, err := server.FormatSSEvent(event1, msg)
	if err != nil {
		t.Error(err)
	}
	expected := "event: result\ndata: <p>msg1</p>\n\n"
	if formatedEvent != expected {
		t.Errorf("expected: %v \nobtained: %v", expected, formatedEvent)
	}
	formatedEvent, err = server.FormatSSEvent(event2, msg)
	if err != nil {
		t.Error(err)
	}
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
	formatter := func(event string, data string) (string, error) {
		formated := event + ":" + data
		return formated, nil
	}
	server.EventStream(rr, testChannel, formatter)
	res := rr.Body.String()
	expected := "result:msg1result:msg2finished:msg2"
	if res != expected {
		t.Errorf("expected: \n%s\nobtained: \n%s", expected, res)
	}
}
