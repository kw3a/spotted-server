package shared

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
	ID          string
	Title       string
	Company     string
	StartDate   string
	EndDate     string
	Description string
}
type EducationEntry struct {
	ID					string
	Degree      string
	Institution string
	StartDate   string
	EndDate     string
}

type Company struct {
	ID          string
	Name        string
	Description string
	Website     string
	ImageURL    string
}
type CompanyQueryParams struct {
	UserID string
	Query  string
}
type JobQueryParams struct {
	Query string
}
