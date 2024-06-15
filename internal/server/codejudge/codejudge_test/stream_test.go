package codejudgetest

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

func TestToString(t *testing.T) {
	res := codejudge.Result{
		Pending:           5,
		Accepted:          1,
		WrongAnswer:       2,
		TimeLimitExceeded: 0,
		RuntimeErrors:     4,
	}
	expected := "Pending: 5, Accepted: 1, Wrong Answer: 2, Runtime Errors: 4"
	if res.ToString() != expected {
		t.Errorf("expected: \n%s\ngot: \n%s", expected, res.ToString())
	}
}

func getTestTokens(length int) []string{
	res := []string{}
	for i := 1; i <= length; i++ {
		res = append(res, fmt.Sprintf("token%d", i))
  }
	return res
}

func TestRegister(t *testing.T) {
	s := codejudge.NewStream()
	tokens := getTestTokens(3)
	err := s.Register("abcde", tokens, 1*time.Second)
	if err != nil {
		t.Error(err)
	}
	err = s.Register("fghij", tokens, 1*time.Second)
	if err != nil {
		t.Error(err)
	}
	err = s.Register("klmno", tokens, 1*time.Second)
	if err != nil {
		t.Error(err)
	}
	err = s.Register("abcde", tokens, 1*time.Second)
	if err == nil {
		t.Error(err)
	}
}

func TestGetTopic(t *testing.T) {
	s := codejudge.NewStream()
	tokens := getTestTokens(3)
	err := s.Register("abcde", tokens, 50*time.Millisecond)
	if err != nil {
		t.Error(err)
	}
	topic, err := s.GetTopic("abcde")
	if err != nil {
		t.Error(err)
	}
	if topic == nil {
		t.Error(err)
	}
	topic, err = s.GetTopic("fghij")
	if err == nil {
		t.Error(err)
	}
	if topic != nil {
		t.Error(err)
	}
	<-time.After(100 * time.Millisecond)
	topic, err = s.GetTopic("abcde")
	if err == nil {
		t.Error(err)
	}
	if topic != nil {
		t.Error(err)
	}
}

func TestUpdateInvalidToken(t *testing.T) {
	tokens := getTestTokens(1)
	topic := codejudge.NewTopic(tokens)
	err := topic.Update("invalid token", "Accepted")
	if err == nil {
		t.Error(err)
	}
}

func TestUpdateUsedToken(t *testing.T) {
	token := getTestTokens(2)
	topic := codejudge.NewTopic(token)
	err := topic.Update("token1", "Accepted")
	if err != nil {
		t.Error(err)
	}
	err = topic.Update("token1", "Wrong Answer")
	if err == nil {
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	tokens := getTestTokens(2)
	topic := codejudge.NewTopic(tokens)
	err := topic.Update("token1", "Accepted")
	if err != nil {
		t.Error(err)
	}
	err = topic.Update("token2", "invalid input")
	if err == nil {
		t.Error(err)
	}
	err = topic.Update("token2", "some Error")
	if err != nil {
		t.Error(err)
	}
	err = topic.Update("token3", "Accepted")
	if err == nil {
		t.Error(err)
	}
}

func TestListenBeforeEmit(t *testing.T) {
	tokens := getTestTokens(2)
	log.Println("lenght of tokens: ", len(tokens))
	topic := codejudge.NewTopic(tokens)
	ch := topic.Listen()
	go func() {
		<-time.After(10 * time.Millisecond)
		err := topic.Update("token1", "Wrong Answer")
		if err != nil {
			t.Error(err)
		}
		err = topic.Update("token2", "Time Limit Exceeded")
		if err != nil {
			t.Error(err)
		}
		fmt.Println("success updates")
	}()
	log.Println("before loop")
	counter := 0
	for msg := range ch {

		log.Println(msg)
		counter++
	}
	log.Println("after loop")
	if counter != 3 {
		t.Error(counter)
	}
}

func TestListenAfterEmit(t *testing.T) {
	tokens := getTestTokens(2)
	topic := codejudge.NewTopic(tokens)
	err := topic.Update("token2", "Accepted")
	if err != nil {
		t.Error(err)
	}
	err = topic.Update("token1", "Wrong Answer")
	if err != nil {
		t.Error(err)
	}
	ch := topic.Listen()
	counter := 0
	for msg := range ch {

		log.Println(msg)
		counter++
	}
	if counter != 1 {
		t.Error(counter)
	}
}

func TestListenInTheMiddle(t *testing.T) {
	tokens := getTestTokens(3)
	log.Println("lenght of tokens: ", len(tokens))
	topic := codejudge.NewTopic(tokens)
	err := topic.Update("token2", "Valid Error")
	if err != nil {
		t.Error(err)
	}
	err = topic.Update("token1", "Accepted")
	if err != nil {
		t.Error(err)
	}
	ch := topic.Listen()
	go func() {
		timer := time.After(10 * time.Millisecond)
		<-timer
		err = topic.Update("token3", "invalid update")
		if err == nil {
			t.Error(err)
		}
		err = topic.Update("token3", "Wrong Answer")
		if err != nil {
			t.Error(err)
		}
	}()
	counter := 0
	fmt.Println("loop start")
	for msg := range ch {
		log.Println(msg)
		counter++
	}
	fmt.Println("loop end")
	if counter != 2 {
		t.Error(counter)
	}

}

func TestListenTwiceAfterEmit(t *testing.T) {
	// listening starts later than transmission
	tokens := getTestTokens(1)
	topic := codejudge.NewTopic(tokens)
	err := topic.Update("token1", "Accepted")
	if err != nil {
		t.Error(err)
	}
	ch1 := topic.Listen()
	counter := 0
	for msg := range ch1 {
		log.Println(msg)
		counter++
	}
	if counter != 1 {
		t.Error(counter)
	}
	ch2 := topic.Listen()
	counter = 0
	for msg := range ch2 {
		log.Println(msg)
		counter++
	}
	if counter != 1 {
		t.Error(counter)
	}
}

func TestEmit(t *testing.T) {
	tokens := getTestTokens(0)
	topic := codejudge.NewTopic(tokens)
	go func() {
		topic.Emit()
	}()
	counter := 0
	for msg := range topic.MessageChannel {
		log.Println(msg)
		counter++
	}
	if counter != 1 {
		t.Error(counter)
	}
}
