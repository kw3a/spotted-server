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

func getXAuthToken() string {
	return "test_token"
}

func getRapidAPIKey() string {
	return "test_rapidapi_token"
}

func getRapidAPIHost() string {
	return "judge0.p.rapidapi.com"
}

func getXAuthHeaders() []Judge0Header {
	return []Judge0Header{
		{
			Name:  "X-Auth-Token",
			Value: getXAuthToken(),
		},
	}
}

func getRapidAPIHeaders() []Judge0Header {
	return []Judge0Header{
		{
			Name:  "x-rapidapi-key",
			Value: getRapidAPIKey(),
		},
		{
			Name:  "x-rapidapi-host",
			Value: getRapidAPIHost(),
		},
	}
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

func getJudgeXAuth() *Judge0 {
	return &Judge0{
		CallbackURL: "http://localhost:42069/api/submissions/",
		JudgeURL:    "",
		Headers:     getXAuthHeaders(),
	}
}

func getJudgeRapidAPI() *Judge0 {
	return &Judge0{
		CallbackURL: "http://localhost:42069/api/submissions/",
		JudgeURL:    "",
		Headers:     getRapidAPIHeaders(),
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

func testServer() *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x_auth_token := r.Header.Get("X-Auth-Token")
		x_rapidapi_key := r.Header.Get("x-rapidapi-key")
		x_rapidapi_host := r.Header.Get("x-rapidapi-host")
		if x_auth_token != getXAuthToken() && x_rapidapi_key != getRapidAPIKey() && x_rapidapi_host != getRapidAPIHost() {
			RespondWithError(w, 400, "invalid token")
			return
		}
		base_64 := r.URL.Query().Get("base64_encoded")
		if base_64 != "true" {
			RespondWithError(w, 400, "base64 not encoded")
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

func TestJsonFormatEmptyTCs(t *testing.T) {
	dbTestCases := []TestCase{}
	submission := getSubmission()
	judge := getJudgeXAuth()
	_, err := JsonFormat(dbTestCases, submission, judge.CallbackURL)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestJsonFormat(t *testing.T) {
	submission := getSubmission()
	judge := getJudgeXAuth()
	dbTestCases := getTestCases()
	body, err := JsonFormat(dbTestCases, submission, judge.CallbackURL)
	if err != nil {
		t.Error(err)
	}
	if body == nil {
		t.Error("Expected body, got nil")
	}
}

func TestComposeUrlEmpty(t *testing.T) {
	_, err := ComposeUrl("", "submissions/batch")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	judgeURL := getJudgeURL()
	_, err = ComposeUrl(judgeURL, "")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestComposeUrl(t *testing.T) {
	judgeUrl := "http://localhost:42069"
	judgeUrl = getJudgeURL()
	url, err := ComposeUrl(judgeUrl, "submissions/batch")
	if err != nil {
		t.Error(err)
	}
	expectedUrl := judgeUrl + "/submissions/batch?" + "base64_encoded=true"
	if url.String() != expectedUrl {
		t.Errorf("Expected %s, got %s", expectedUrl, url.String())
	}
}

func TestSendRequestBadBody(t *testing.T) {
	body := []byte{}
	URL := url.URL{
		Scheme: "http",
		Host:   "localhost:42069",
		Path:   "test",
	}
	_, err := SendRequest(body, URL, []Judge0Header{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestBadURL(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	body, err := JsonFormat(dbTestCases, submission, "http://localhost:42069/callback")
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
	_, err = SendRequest(body, URL, []Judge0Header{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestUnexpectedStatusCode(t *testing.T) {
	submission := getSubmission()
	dbTestCases := getTestCases()
	judge := getJudgeXAuth()
	body, err := JsonFormat(dbTestCases, submission, judge.CallbackURL)
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
	_, err = SendRequest(body, URL, []Judge0Header{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendRequestXAuth(t *testing.T) {
	judgeTCs := []Judge0TC{
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
	URLStr := testServer.URL + "?base64_encoded=true"
	URL, err := url.Parse(URLStr)
	if err != nil {
		t.Error(err)
	}
	tokens, err := SendRequest(body, *URL, getXAuthHeaders())
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestSendRequestRapidAPI(t *testing.T) {
	judgeTCs := []Judge0TC{
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
	URLStr := testServer.URL + "?base64_encoded=true"
	URL, err := url.Parse(URLStr)
	if err != nil {
		t.Error(err)
	}
	tokens, err := SendRequest(body, *URL, getRapidAPIHeaders())
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestSendEmptyTC(t *testing.T) {
	judge := getJudgeXAuth()
	submission := getSubmission()
	_, err := judge.Send([]TestCase{}, submission)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendEmptySubmission(t *testing.T) {
	judge := getJudgeXAuth()
	dbTestCases := getTestCases()
	_, err := judge.Send(dbTestCases, Submission{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendBadJudgeURL(t *testing.T) {
	judge := getJudgeXAuth()
	submission := getSubmission()
	dbTestCases := getTestCases()
	judge.JudgeURL = ""
	_, err := judge.Send(dbTestCases, submission)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSendXAuth(t *testing.T) {
	judge := getJudgeXAuth()
	submission := getSubmission()
	dbTestCases := getTestCases()
	testServer := testServer()
	defer testServer.Close()
	judge.JudgeURL = testServer.URL
	tokens, err := judge.Send(dbTestCases, submission)
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf(`Expected 2 tokens, got %d`, len(tokens))
	}
}

func TestSendRapidAPI(t *testing.T) {
	judge := getJudgeRapidAPI()
	submission := getSubmission()
	dbTestCases := getTestCases()
	testServer := testServer()
	defer testServer.Close()
	judge.JudgeURL = testServer.URL
	tokens, err := judge.Send(dbTestCases, submission)
	if err != nil {
		t.Error(err)
	}
	if len(tokens) != 2 {
		t.Errorf(`Expected 2 tokens, got %d`, len(tokens))
	}
}
