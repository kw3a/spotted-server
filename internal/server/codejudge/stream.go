package codejudge

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Accepted          int32
	WrongAnswer       int32
	RuntimeErrors     int32
	TimeLimitExceeded int32
	Pending           int
}

func (r *Result) ToString() string {
	res := fmt.Sprintf("En cola: %d", r.Pending)
	if r.Accepted > 0 {
		res += fmt.Sprintf(", Aceptado: %d", r.Accepted)
	}
	if r.WrongAnswer > 0 {
		res += fmt.Sprintf(", Respuesta equivocada: %d", r.WrongAnswer)
	}
	if r.RuntimeErrors > 0 {
		res += fmt.Sprintf(", Error: %d", r.RuntimeErrors)
	}
	if r.TimeLimitExceeded > 0 {
		res += fmt.Sprintf(", Tiempo límite excedido: %d", r.TimeLimitExceeded)
	}
	return res
}

type Stream struct {
	topics map[string]*Topic
	mu     sync.Mutex
}

type Topic struct {
	MessageChannel chan string
	listening      bool
	result         Result
	tokens         map[string]bool
	mu             sync.Mutex
}

func NewStream() *Stream {
	return &Stream{
		topics: make(map[string]*Topic),
		mu:     sync.Mutex{},
	}
}

func NewTopic(tokens []string) *Topic {
	return &Topic{
		MessageChannel: make(chan string),
		listening:      false,
		result:         Result{Pending: len(tokens)},
		tokens:         TokenMap(tokens),
		mu:             sync.Mutex{},
	}
}

func TokenMap(tokens []string) map[string]bool {
	res := make(map[string]bool)
	for _, token := range tokens {
		res[token] = false
	}
	return res
}
func (s *Stream) Register(name string, tokens []string, duration time.Duration) error {
	s.mu.Lock()
	_, ok := s.topics[name]
	s.mu.Unlock()
	if ok {
		return fmt.Errorf("topic %s already exists", name)
	}
	topic := NewTopic(tokens)
	s.mu.Lock()
	s.topics[name] = topic
	s.mu.Unlock()
	go s.automaticDelete(name, topic, duration)
	return nil
}

func (s *Stream) automaticDelete(name string, topic *Topic, duration time.Duration) {
	<-time.After(duration)
	s.mu.Lock()
	if topic.listening {
		close(topic.MessageChannel)
	}
	delete(s.topics, name)
	log.Printf("Topic: %s deleted\n", name)
	log.Println("new length: ", len(s.topics))
	s.mu.Unlock()
}

func (s *Stream) GetTopic(name string) (*Topic, error) {
	s.mu.Lock()
	t, ok := s.topics[name]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("topic %s not found", name)
	}
	return t, nil
}

func (t *Topic) Listen() chan string {
	t.mu.Lock()
	t.listening = true
	t.MessageChannel = make(chan string)
	t.mu.Unlock()
	go func() {
		t.Emit()
	}()

	return t.MessageChannel
}

func (t *Topic) Update(token, status string) error {
	if t.result.Pending == 0 {
		return fmt.Errorf("topic is finished")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	used, ok := t.tokens[token]
	if !ok {
		return fmt.Errorf("token doesn't exists: %s", token)
	}
	if used {
		return fmt.Errorf("token is already in use: %s", token)
	}
	if strings.Contains(status, "Error") {
		t.result.RuntimeErrors++
	} else {
		switch status {
		case "Accepted":
			t.result.Accepted++
		case "Wrong Answer":
			t.result.WrongAnswer++
		case "Time Limit Exceeded":
			t.result.TimeLimitExceeded++
		default:
			return fmt.Errorf("invalid status: %s", status)
		}
	}
	t.result.Pending--
	t.tokens[token] = true
	if t.listening {
		t.Emit()
	}
	return nil
}

func (t *Topic) Emit() {
	t.MessageChannel <- t.result.ToString()
	if t.result.Pending == 0 {
		close(t.MessageChannel)
		t.listening = false
	}
}
