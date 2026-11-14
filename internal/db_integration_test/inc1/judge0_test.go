package inc1test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/stretchr/testify/require"
)

type judgeCallback struct {
	Input        shared.CallbackJsonInput
	SubmissionID string
	TestCaseID   string
}

// Records the callback after the real server handler decodes it. MySQL is
// outside the server–Judge0 boundary and is covered by the database suite.
type judgeCallbackStorage struct{ received chan judgeCallback }

func (s *judgeCallbackStorage) UpdateTestCaseResult(ctx context.Context, input shared.CallbackJsonInput, submissionID, testCaseID string) error {
	select {
	case s.received <- judgeCallback{input, submissionID, testCaseID}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type judgeFixture struct {
	Judge    codejudge.Judge0
	Stream   *codejudge.Stream
	Received chan judgeCallback
	Ready    chan struct{}
	release  sync.Once
	Ctx      context.Context
}

func newJudgeFixture(t *testing.T) *judgeFixture {
	t.Helper()
	endpoint := os.Getenv("JUDGE0_INTEGRATION_URL")
	if endpoint == "" {
		t.Skip("JUDGE0_INTEGRATION_URL required; use internal/db_integration_test/run-judge0.sh")
	}
	callbackHost := os.Getenv("JUDGE0_INTEGRATION_CALLBACK_HOST")
	require.NotEmpty(t, callbackHost, "callback host must be reachable from Judge0 workers")
	token := os.Getenv("JUDGE0_INTEGRATION_TOKEN")
	require.NotEmpty(t, token, "authentication token is required")
	// The production client uses http.DefaultClient. Bound its requests for these
	// sequential integration tests and restore it before leaving each case.
	previous := http.DefaultClient
	bounded := *previous
	bounded.Timeout = 15 * time.Second
	http.DefaultClient = &bounded
	t.Cleanup(func() { http.DefaultClient = previous })
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	f := &judgeFixture{Stream: codejudge.NewStream(), Received: make(chan judgeCallback, 32), Ready: make(chan struct{}), Ctx: ctx}
	router := chi.NewRouter()
	handler := quizes.CreateCallbackHandler(&judgeCallbackStorage{f.Received}, f.Stream, shared.Decode[shared.CallbackJsonInput], quizes.GetCallbackURLParamsInput)
	router.Put("/api/submissions/{submissionID}/tc/{testCaseID}", func(w http.ResponseWriter, r *http.Request) {
		// A very fast worker can finish before Send returns its tokens. Keep the
		// callback pending until the real stream has been registered and consumed.
		select {
		case <-f.Ready:
			handler(w, r)
		case <-ctx.Done():
			http.Error(w, "test timed out", http.StatusGatewayTimeout)
		}
	})
	listener, err := net.Listen("tcp", net.JoinHostPort(callbackHost, "0"))
	require.NoError(t, err)
	server := httptest.NewUnstartedServer(router)
	require.NoError(t, server.Listener.Close())
	server.Listener = listener
	server.Start()
	t.Cleanup(func() { f.release.Do(func() { close(f.Ready) }); server.Close() })
	f.Judge = codejudge.NewJudge0(endpoint, server.URL+"/api/submissions/", []codejudge.Judge0Header{{Name: "X-Auth-Token", Value: token}})
	return f
}

func (f *judgeFixture) send(t *testing.T, cases []codejudge.TestCase, source string, language int32) (string, []string, <-chan string) {
	t.Helper()
	submission := uuid.NewString()
	tokens, err := f.Judge.Send(cases, codejudge.Submission{ID: submission, Src: source, LanguageID: language})
	require.NoError(t, err)
	require.Len(t, tokens, len(cases))
	unique := map[string]bool{}
	for _, token := range tokens {
		require.NoError(t, uuid.Validate(token))
		require.False(t, unique[token], "tokens must be unique")
		unique[token] = true
	}
	require.NoError(t, f.Stream.Register(submission, tokens, 2*time.Minute))
	listener, err := f.Stream.Listen(submission)
	require.NoError(t, err)
	finished := make(chan string, 1)
	go func() {
		last := ""
		for message := range listener {
			last = message
		}
		finished <- last
	}()
	f.release.Do(func() { close(f.Ready) })
	return submission, tokens, finished
}

func (f *judgeFixture) callbacks(t *testing.T, submission string, cases []codejudge.TestCase, tokens []string) map[string]shared.CallbackJsonInput {
	t.Helper()
	expected := map[string]string{}
	for i, c := range cases {
		expected[c.ID] = tokens[i]
	}
	received := map[string]shared.CallbackJsonInput{}
	for len(received) < len(cases) {
		select {
		case callback := <-f.Received:
			require.Equal(t, submission, callback.SubmissionID)
			token, ok := expected[callback.TestCaseID]
			require.True(t, ok, "unexpected test case")
			require.Equal(t, token, callback.Input.Token)
			_, duplicate := received[callback.TestCaseID]
			require.False(t, duplicate, "duplicate callback")
			received[callback.TestCaseID] = callback.Input
		case <-f.Ctx.Done():
			t.Fatalf("Judge0 callbacks incomplete: received %d/%d: %v", len(received), len(cases), f.Ctx.Err())
		}
	}
	return received
}

func (f *judgeFixture) final(t *testing.T, finished <-chan string) string {
	t.Helper()
	select {
	case result := <-finished:
		return result
	case <-f.Ctx.Done():
		t.Fatal("callback stream did not finish before the deadline")
		return ""
	}
}

func TestJudge0Verdicts(t *testing.T) {
	for _, tc := range []struct {
		name, source, status, output, summary string
		language                              int32
	}{
		{"acceptedCpp", `#include <iostream>
int main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<"\n";}`, "Accepted", "5\n", "Aceptado: 1", 54},
		{"acceptedJava", `public class Main { public static void main(String[] args) { java.util.Scanner s=new java.util.Scanner(System.in); System.out.println(s.nextInt()+s.nextInt()); } }`, "Accepted", "5\n", "Aceptado: 1", 62},
		{"wrongAnswer", `#include <iostream>
int main(){std::cout<<"wrong\n";}`, "Wrong Answer", "wrong\n", "Respuesta equivocada: 1", 54},
		{"compilationError", `int main( { this is invalid C++;`, "Compilation Error", "", "Error: 1", 54},
		{"runtimeError", `int main(){return 1;}`, "Runtime Error (NZEC)", "", "Error: 1", 54},
		{"timeLimitExceeded", `int main(){volatile unsigned long long n=0;for(;;){n++;}}`, "Time Limit Exceeded", "", "Tiempo límite excedido: 1", 54},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newJudgeFixture(t)
			cases := []codejudge.TestCase{{ID: uuid.NewString(), Input: "2 3\n", ExpectedOutput: "5\n", TimeLimit: 1, MemoryLimit: 262144}}
			submission, tokens, finished := f.send(t, cases, tc.source, tc.language)
			callbacks := f.callbacks(t, submission, cases, tokens)
			result := callbacks[cases[0].ID]
			require.Equal(t, tc.status, result.Status.Description)
			require.Equal(t, tc.output, result.Stdout, "callback output must be decoded from base64 by the server")
			require.False(t, result.Time.IsNegative())
			require.GreaterOrEqual(t, result.Memory, int32(0))
			if tc.status == "Compilation Error" {
				require.NotEmpty(t, result.CompileOutput)
			}
			final := f.final(t, finished)
			require.Contains(t, final, "En cola: 0")
			require.Contains(t, final, tc.summary)
		})
	}
}

func TestJudge0BatchCallbacks(t *testing.T) {
	f := newJudgeFixture(t)
	cases := []codejudge.TestCase{
		{ID: uuid.NewString(), Input: "2 3\n", ExpectedOutput: "5\n", TimeLimit: 1, MemoryLimit: 262144},
		{ID: uuid.NewString(), Input: "10 20\n", ExpectedOutput: "999\n", TimeLimit: 1, MemoryLimit: 262144},
	}
	source := `#include <iostream>
int main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<"\n";}`
	submission, tokens, finished := f.send(t, cases, source, 54)
	callbacks := f.callbacks(t, submission, cases, tokens)
	require.Equal(t, "Accepted", callbacks[cases[0].ID].Status.Description)
	require.Equal(t, "5\n", callbacks[cases[0].ID].Stdout)
	require.Equal(t, "Wrong Answer", callbacks[cases[1].ID].Status.Description)
	require.Equal(t, "30\n", callbacks[cases[1].ID].Stdout)
	final := f.final(t, finished)
	require.Contains(t, final, "En cola: 0")
	require.Contains(t, final, "Aceptado: 1")
	require.Contains(t, final, "Respuesta equivocada: 1")
}

func TestJudge0RejectsInvalidAuthentication(t *testing.T) {
	f := newJudgeFixture(t)
	f.Judge.Headers = []codejudge.Judge0Header{{Name: "X-Auth-Token", Value: "invalid-" + uuid.NewString()}}
	cases := []codejudge.TestCase{{ID: uuid.NewString(), TimeLimit: 1, MemoryLimit: 262144}}
	tokens, err := f.Judge.Send(cases, codejudge.Submission{ID: uuid.NewString(), Src: "int main(){}", LanguageID: 54})
	require.Error(t, err)
	require.Contains(t, err.Error(), "status code: 401")
	require.Empty(t, tokens)
}

// Compile-time interface check documents the in-memory callback observation point.
var _ quizes.CallbackStorage = (*judgeCallbackStorage)(nil)
