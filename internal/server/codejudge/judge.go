package codejudge

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Judge0TC struct {
	Src            string  `json:"source_code"`
	LanguageID     int32   `json:"language_id"`
	Memory_limit   int32   `json:"memory_limit"`
	Time_limit     float64 `json:"cpu_time_limit"`
	Input          string  `json:"stdin"`
	ExpectedOutput string  `json:"expected_output"`
	CallbackURL    string  `json:"callback_url"`
}

type JudgeSubmission struct {
	TestsCases []Judge0TC `json:"submissions"`
}

type Token struct {
	Token string `json:"token"`
}
type Submission struct {
	ID         string
	Src        string
	LanguageID int32
}

type TestCase struct {
	ID             string
	TimeLimit      float64
	MemoryLimit    int32
	Input          string
	ExpectedOutput string
}

type Judge0 struct {
	CallbackURL string
	JudgeURL    string
	AuthToken   string
}

func NewJudge0(judgeURL, authToken, callbackURL string) Judge0 {
	return Judge0{
		CallbackURL: callbackURL,
		JudgeURL:    judgeURL,
		AuthToken:   authToken,
	}
}

func (j *Judge0) Send(testCases []TestCase, submission Submission) ([]string, error) {
	body, err := JsonFormat(testCases, submission, j.CallbackURL)
	if err != nil {
		return []string{}, err
	}
	URL, err := ComposeUrl(j.JudgeURL, "submissions/batch", j.AuthToken)
	if err != nil {
		return []string{}, err
	}
	return SendRequest(body, URL)
}

func JsonFormat(dbTestCases []TestCase, submission Submission, callbackURL string) ([]byte, error) {
	if len(dbTestCases) == 0 {
		return nil, errors.New("empty database test cases")
	}
	judgeTCs := []Judge0TC{}
	for _, dbTestCase := range dbTestCases {
		current := Judge0TC{
			Src:            encode(submission.Src),
			LanguageID:     submission.LanguageID,
			Memory_limit:   dbTestCase.MemoryLimit,
			Time_limit:     dbTestCase.TimeLimit,
			Input:          encode(dbTestCase.Input),
			ExpectedOutput: encode(dbTestCase.ExpectedOutput),
			CallbackURL:    callbackURL + submission.ID + "/tc/" + dbTestCase.ID,
		}
		judgeTCs = append(judgeTCs, current)
	}
	judgeSubmission := JudgeSubmission{
		TestsCases: judgeTCs,
	}
	jsonJudgeTC, err := json.Marshal(judgeSubmission)
	if err != nil {
		return []byte{}, err
	}
	return jsonJudgeTC, nil
}

func SendRequest(body []byte, URL url.URL) ([]string, error) {
	bodyReader := bytes.NewReader(body)
	resp, err := http.Post(URL.String(), "application/json", bodyReader)
	if err != nil {
		return []string{}, errors.New("Post submission failed: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return []string{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	successResp := []Token{}
	err = json.NewDecoder(resp.Body).Decode(&successResp)
	if err != nil {
		return []string{}, err
	}
	res := []string{}
	for _, token := range successResp {
		res = append(res, token.Token)
	}
	return res, nil
}

func ComposeUrl(host, path, authToken string) (url.URL, error) {
	if host == "" {
		return url.URL{}, errors.New("empty host")
	}
	if path == "" {
		return url.URL{}, errors.New("empty path")
	}
	if authToken == "" {
		return url.URL{}, errors.New("empty auth token")
	}
	parsedHost, err := url.Parse(host)
	if err != nil {
		return url.URL{}, err
	}
	u := url.URL{
		Scheme: parsedHost.Scheme,
		Host:   parsedHost.Host,
		Path:   path,
	}
	q := u.Query()
	q.Set("base64_encoded", "true")
	q.Set("X-Auth-Token", authToken)
	u.RawQuery = q.Encode()
	return u, nil
}

func encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}
