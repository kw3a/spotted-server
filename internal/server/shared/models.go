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
	Page   int32
	UserID string
	Query  string
}
type OfferQueryParams struct {
	Page      int32
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
	TimeLimit   int32
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
	Time         int32
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

type StrokeWindow struct {
	ID           string
	StrokeAmount int32
	UdMean       int32
	UdStdDev     int32
	Du1Mean      int32
	Du1StdDev    int32
	Du2Mean      int32
	Du2StdDev    int32
	DdMean       int32
	DdStdDev     int32
	UuMean       int32
	UuStdDev     int32
}
