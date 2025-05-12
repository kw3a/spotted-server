package shared

import (
	"time"
)

type User struct {
	ID          string
	Name        string
	Cell        string
	ImageURL    string
	Email       string
	Description string
}

type SkillEntry struct {
	Name string
	ID   string
}
type Link struct {
	URL  string
	Name string
	ID   string
}
type ExperienceEntry struct {
	ID           string
	Title        string
	Company      string
	StartDate    string
	EndDate      string
	Description  string
	TimeInterval string
}
type EducationEntry struct {
	ID           string
	Degree       string
	Institution  string
	StartDate    string
	EndDate      string
	TimeInterval string
}

type Company struct {
	ID          string
	UserID      string
	Name        string
	Description string
	Website     string
	ImageURL    string
}
type CompanyQueryParams struct {
	UserID string
	Query  string
}
type OfferQueryParams struct {
	UserID    string
	CompanyID string
	Query     string
}
type Language struct {
	ID          int32
	Name        string
	DisplayName string
}

type Offer struct {
	ID              string
	Title           string
	About           string
	Requirements    string
	Benefits        string
	Status          int32
	CompanyName     string
	CompanyID       string
	CompanyImageURL string
	MinWage         int32
	MaxWage         int32
	RelativeTime    string
}

type Quiz struct {
	ID        string
	Duration  int32
	Languages []int32
}

type Problem struct {
	ID          string
	Title       string
	Description string
	MemoryLimit int32
	TimeLimit   float64
	TestCases   []TestCase
	Examples    []Example
}

type Example struct {
	ID        string
	Input     string
	Output    string
	ProblemID string
}

type Score struct {
	AcceptedTestCases int
	TotalTestCases    int
}

type Submission struct {
	ID                string
	Src               string
	AcceptedTestCases int
	ParticipationID   string
	LanguageID        int32
	ProblemID         string
	Language          string
}

type Participation struct {
	ID           string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RelativeTime string
}

type TestCase struct {
	ID        string
	Input     string
	Output    string
	ProblemID string
}

type TestCaseResult struct {
	ID           string
	Output       string
	Status       string
	Time         int64
	Memory       int32
	SubmissionID string
	TestCaseID   string
}

type ExecutedTestCase struct {
	TestCase TestCase
	Result   TestCaseResult
}

type Application struct {
	Applicant     User
	Participation Participation
	Summary       []Summary
}

type Summary struct {
	Title      string
	Score      Score
	Submission Submission
	Results    []TestCaseResult
}
