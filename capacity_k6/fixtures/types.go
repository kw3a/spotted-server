package main

const datasetVersion = 4

type Sample struct{ Input, Output string }
type Algorithm struct {
	Key, Title, Description string
	PartialAccepted         int
	PartialCaseAccepted     [5]bool
	Examples, Cases         []Sample
}
type Verdict struct {
	Accepted          int `json:"accepted"`
	WrongAnswer       int `json:"wrongAnswer"`
	Errors            int `json:"errors"`
	TimeLimitExceeded int `json:"timeLimitExceeded"`
}
type Solution struct {
	Kind     string  `json:"kind"`
	Source   string  `json:"source"`
	Expected Verdict `json:"expected"`
}

type Problem struct {
	ProblemID     string     `json:"problemID"`
	Title         string     `json:"title"`
	LanguageID    string     `json:"languageID"`
	Source        string     `json:"source"`
	TestCaseCount int        `json:"testCaseCount"`
	Solutions     []Solution `json:"solutions"`
}
type Offer struct {
	OfferID   string    `json:"offerID"`
	CompanyID string    `json:"companyID"`
	QuizID    string    `json:"quizID"`
	Problems  []Problem `json:"problems"`
}
type Manifest struct {
	Version int     `json:"version"`
	Offers  []Offer `json:"offers"`
}
