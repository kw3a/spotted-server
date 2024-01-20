package server

import "gitlab.com/kw3a/spotted-server/internal/database"

type Problem struct {
	ProblemData
	Languages []Language `json:"languages"`
	Examples  []Example  `json:"examples"`
}

type ProblemData struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Title       string  `json:"title"`
	MemoryLimit int32   `json:"memory_limit"`
	TimeLimit   float64 `json:"time_limit"`
	QuizID      string  `json:"quiz_id"`
}

type Language struct {
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Version int32  `json:"version"`
}

type Example struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func databaseProblemToProblem(problem []database.GetProblemRow) Problem {
	if len(problem) == 0 {
		return Problem{}
	}
	languages := []Language{}
	examples := []Example{}
	for _, row := range problem {
		if row.Input == "" {
			languages = append(languages, Language{
				ID:      row.LanguageID,
				Name:    row.LanguageName,
				Version: row.LanguageVersion,
			})
			continue
		}
		examples = append(examples, Example{
			Input:  row.Input,
			Output: row.Output,
		})
	}
	return Problem{
		ProblemData: ProblemData{
			ID:          problem[0].ID,
			Description: problem[0].Description,
			Title:       problem[0].Title,
			MemoryLimit: problem[0].MemoryLimit,
			TimeLimit:   problem[0].TimeLimit,
			QuizID:      problem[0].QuizID,
		},
		Languages: languages,
		Examples:  examples,
	}
}

func databaseProblemsToProblems(problems []database.GetProblemsRow) []Problem {
	if len(problems) == 0 {
		return nil
	}

	var result []Problem
	var languages []Language
	var examples []Example
	currentProblemID := problems[0].ID
	currentProblem := ProblemData{
		ID:          problems[0].ID,
		Description: problems[0].Description,
		Title:       problems[0].Title,
		MemoryLimit: problems[0].MemoryLimit,
		TimeLimit:   problems[0].TimeLimit,
		QuizID:      problems[0].QuizID,
	}

	for _, row := range problems {
		if row.ID != currentProblemID {
			result = append(result, Problem{
				ProblemData: currentProblem,
				Languages:   languages,
				Examples:    examples,
			})
			currentProblemID = row.ID
			languages = nil
			examples = nil
			currentProblem = ProblemData{
				ID:          row.ID,
				Description: row.Description,
				Title:       row.Title,
				MemoryLimit: row.MemoryLimit,
				TimeLimit:   row.TimeLimit,
				QuizID:      row.QuizID,
			}
		}

		if row.Input == "" {
			languages = append(languages, Language{
				ID:      row.LanguageID,
				Name:    row.LanguageName,
				Version: row.LanguageVersion,
			})
		} else {
			examples = append(examples, Example{
				Input:  row.Input,
				Output: row.Output,
			})
		}
	}
	result = append(result, Problem{
		ProblemData: currentProblem,
		Languages:   languages,
		Examples:    examples,
	})
	return result
}
