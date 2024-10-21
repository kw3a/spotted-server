package codejudge

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func getAuthToken() string {
	return "test_token"
}

func getJudgeURL() string {
	return "http://localhost:42069"
}

func getSubmission() Submission {
	pythonID := 71
	return Submission{
		ID:         uuid.NewString(),
		Src:        "print('hello')",
		LanguageID: int32(pythonID),
	}
}

func getJudge() *Judge0 {
	return &Judge0{
		CallbackURL: "http://localhost:42069/api/submissions/",
		JudgeURL:    "",
		AuthToken:   "test_token",
	}
}

func getTestCases() []TestCase {
	return []TestCase{
		{
			ID:             uuid.NewString(),
			TimeLimit:      1,
			MemoryLimit:    2048,
			Input:          "",
			ExpectedOutput: "Hello",
		},
		{
			ID:             uuid.NewString(),
			TimeLimit:      1,
			MemoryLimit:    2048,
			Input:          "1",
			ExpectedOutput: "2",
		},
	}
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}
func TestComposeUrlEmpty(t *testing.T) {
	auth_token := getAuthToken()
	_, err := ComposeUrl("", "submissions/batch", auth_token)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	judgeURL := getJudgeURL()
	_, err = ComposeUrl(judgeURL, "", auth_token)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	_, err = ComposeUrl(judgeURL, "submissions/batch", "")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestComposeUrl(t *testing.T) {
	judgeUrl := "http://localhost:42069"
	authToken := getAuthToken()
	judgeUrl = getJudgeURL()
	url, err := ComposeUrl(judgeUrl, "submissions/batch", authToken)
	if err != nil {
		t.Error(err)
	}
	expectedUrl := judgeUrl + "/submissions/batch?" + "X-Auth-Token=" + authToken + "&" + "base64_encoded=true"
	if url.String() != expectedUrl {
		t.Errorf("Expected %s, got %s", expectedUrl, url.String())
	}
}

func TestDBTCsToJsonEmpty(t *testing.T) {
	dbTestCases := []TestCase{}
	submission := getSubmission()
	judge := getJudge()
	_, err := DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDBTCsToJson(t *testing.T) {
	submission := getSubmission()
	judge := getJudge()
	dbTestCases := getTestCases()
	body, err := DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
	if err != nil {
		t.Error(err)
	}
	if body == nil {
		t.Error("Expected body, got nil")
	}
}

func TestSendRequestBadBody(t *testing.T) {
	body := []byte{}
	URL := url.URL{
		Scheme: "http",
		Host:   "localhost:42069",
		Path:   "test",
	}
	_, err := SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestBadURL(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	body, err := DBTCsToJson(dbTestCases, submission, "http://localhost:42069/callback")
	if err != nil {
		t.Error(err)
	}
	if body == nil {
		t.Error("Expected body, got nil")
	}
	URL := url.URL{
		Scheme: "http",
		Host:   "randomdomain.zxczxd",
	}
	_, err = SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestUnexpectedStatusCode(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	judge := getJudge()
	body, err := DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
	if err != nil {
		t.Error(err)
	}
	if body == nil {
		t.Error("Expected body, got nil")
	}
	URL := url.URL{
		Scheme: "http",
		Host:   "localhost:42069",
		Path:   "test",
	}
	_, err = SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequest(t *testing.T) {
	judgeTCs := []JudgeTestCase{
		{},
		{},
	}
	payload := JudgeSubmission{
		TestsCases: judgeTCs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Error(err)
	}
	testServer := testServer()
	defer testServer.Close()
	URLStr := testServer.URL + "?X-Auth-Token=test_token&base64_encoded=true"
	URL, err := url.Parse(URLStr)
	if err != nil {
		t.Error(err)
	}
	tokens, err := SendRequest(body, *URL)
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func testServer() *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("X-Auth-Token")
		base_64 := r.URL.Query().Get("base64_encoded")
		if token != "test_token" || base_64 != "true" {
			RespondWithError(w, 400, "invalid token")
			return
		}
		params := JudgeSubmission{}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			RespondWithError(w, 401, "can't unmarshal params")
			return
		}
		if len(params.TestsCases) != 2 {
			RespondWithError(w, 402, "invalid length")
			return
		}
		payload := []Token{{Token: "token1"}, {Token: "token2"}}
		RespondWithJSON(w, http.StatusCreated, payload)
	}))
	return testServer
}

func TestSendEmptyTC(t *testing.T) {
	judge := getJudge()
	submission := getSubmission()
	_, err := judge.Send([]TestCase{}, submission.ID, submission.Src, submission.LanguageID)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendEmptySubmission(t *testing.T) {
	judge := getJudge()
	dbTestCases := getTestCases()
	_, err := judge.Send(dbTestCases, "", "", 0)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendBadJudgeURL(t *testing.T) {
	judge := getJudge()
	submission := getSubmission()
	dbTestCases := getTestCases()
	judge.JudgeURL = ""
	_, err := judge.Send(dbTestCases, submission.ID, submission.Src, submission.LanguageID)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
func TestSend(t *testing.T) {
	judge := getJudge()
	submission := getSubmission()
	dbTestCases := getTestCases()
	testServer := testServer()
	defer testServer.Close()
	judge.JudgeURL = testServer.URL
	tokens, err := judge.Send(dbTestCases, submission.ID, submission.Src, submission.LanguageID)
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf(`Expected 2 tokens, got %d`, len(tokens))
	}
}
