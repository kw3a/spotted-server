package codejudgetest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	responseparser "github.com/kw3a/spotted-server/internal/http_parser/responseParser"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

func getAuthToken() string {
	return "test_token"
}

func getJudgeURL() string {
	return "http://localhost:42069"
}

func getSubmission() codejudge.Submission {
	pythonID := 71
	return codejudge.Submission{
		ID:         uuid.NewString(),
		Src:        "print('hello')",
		LanguageID: int32(pythonID),
	}
}

func getJudge() *codejudge.Judge0 {
	return &codejudge.Judge0{
		CallbackURL: "http://localhost:42069/api/submissions/",
		JudgeURL:    "",
		AuthToken:   "test_token",
	}
}

func getTestCases() []codejudge.TestCase {
	return []codejudge.TestCase{
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

func TestComposeUrlEmpty(t *testing.T) {
	auth_token := getAuthToken()
	_, err := codejudge.ComposeUrl("", "submissions/batch", auth_token)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	judgeURL := getJudgeURL()
	_, err = codejudge.ComposeUrl(judgeURL, "", auth_token)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	_, err = codejudge.ComposeUrl(judgeURL, "submissions/batch", "")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestComposeUrl(t *testing.T) {
	judgeUrl := "http://localhost:42069"
	authToken := getAuthToken()
	judgeUrl = getJudgeURL()
	url, err := codejudge.ComposeUrl(judgeUrl, "submissions/batch", authToken)
	if err != nil {
		t.Error(err)
	}
	expectedUrl := judgeUrl + "/submissions/batch?" + "X-Auth-Token=" + authToken + "&" + "base64_encoded=true"
	if url.String() != expectedUrl {
		t.Errorf("Expected %s, got %s", expectedUrl, url.String())
	}
}

func TestDBTCsToJsonEmpty(t *testing.T) {
	dbTestCases := []codejudge.TestCase{}
	submission := getSubmission()
	judge := getJudge()
	_, err := codejudge.DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDBTCsToJson(t *testing.T) {
	submission := getSubmission()
	judge := getJudge()
	dbTestCases := getTestCases()
	body, err := codejudge.DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
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
	_, err := codejudge.SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestBadURL(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	body, err := codejudge.DBTCsToJson(dbTestCases, submission, "http://localhost:42069/callback")
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
	_, err = codejudge.SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestUnexpectedStatusCode(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	judge := getJudge()
	body, err := codejudge.DBTCsToJson(dbTestCases, submission, judge.CallbackURL)
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
	_, err = codejudge.SendRequest(body, URL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequest(t *testing.T) {
	judgeTCs := []codejudge.JudgeTestCase{
		{},
		{},
	}
	payload := codejudge.JudgeSubmission{
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
	tokens, err := codejudge.SendRequest(body, *URL)
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
			responseparser.RespondWithError(w, 400, "invalid token")
			return
		}
		params := codejudge.JudgeSubmission{}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			responseparser.RespondWithError(w, 401, "can't unmarshal params")
			return
		}
		if len(params.TestsCases) != 2 {
			responseparser.RespondWithError(w, 402, "invalid length")
			return
		}
		payload := []codejudge.Token{{Token: "token1"}, {Token: "token2"}}
		responseparser.RespondWithJSON(w, http.StatusCreated, payload)
	}))
	return testServer
}

func TestSendEmptyTC(t *testing.T) {
	judge := getJudge()
	submission := getSubmission()
	_, err := judge.Send([]codejudge.TestCase{}, submission.ID, submission.Src, submission.LanguageID)
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
