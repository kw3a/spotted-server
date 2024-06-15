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

func TestFormatServerSentEvent(t *testing.T) {
	msg := "msg1"
	event := "result"
	str, err := server.FormatServerSentEvent(event, msg)
	if err != nil {
		t.Error(err)
	}
	expected := "event: result\ndata: msg1\n\n"
	if str != expected {
		t.Errorf("got: %v \nwant: %v", str, expected)
	}
}

func TestEventStream(t *testing.T) {
	msg := "msg1"
	msg2 := "msg2"
	expected := "event: result\ndata: msg1\n\nevent: result\ndata: msg2\n\nevent: finished\ndata: msg2\n\n"
	testChannel := make(chan string)
	go func() {
		testChannel <- msg
		testChannel <- msg2
		close(testChannel)
	}()
	rr := httptest.NewRecorder()
	server.EventStream(rr, testChannel)
	res := rr.Body.String()
	if res != expected {
		t.Errorf("expected: \n%s\nobtained: \n%s", expected, res)
	}
}
