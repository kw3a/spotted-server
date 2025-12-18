package offers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kw3a/spotted-server/internal/server/shared"
)

const availableLanguages = 6
const memoryLimit = 262144

type OfferRegInput struct {
	Offer    shared.Offer
	Quiz     shared.Quiz
	Problems []shared.Problem
}

type OfferRegJ struct {
	Offer    OfferJ     `json:"offer"`
	Quiz     QuizJ      `json:"quiz"`
	Problems []ProblemJ `json:"problems"`
}

type OfferJ struct {
	Title        string `json:"title"`
	MinWage      string `json:"minWage"`
	MaxWage      string `json:"maxWage"`
	CompanyID    string `json:"companyID"`
	About        string `json:"about"`
	Requirements string `json:"requirements"`
	Benefits     string `json:"benefits"`
}

type QuizJ struct {
	Duration  string   `json:"duration"`
	Languages []string `json:"languages"`
}

type ProblemJ struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	TimeLimit   string      `json:"time_limit"`
	TestCases   []TestCaseJ `json:"test_cases"`
	Examples    []ExampleJ  `json:"examples"`
}

type TestCaseJ struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type ExampleJ struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type OfferRegistrationInputErrors struct {
	CompanyIDError   string
	TitleError       string
	DescriptionError string
	WageError        string
}

func lenError(field string, min int, max int) error {
	return fmt.Errorf("el campo '%s' debe tener entre %d y %d caracteres", field, min, max)
}

func ValidateOffer(o OfferJ) (shared.Offer, error) {
	if len(o.Title) < 10 || len(o.Title) > 60 {
		return shared.Offer{}, lenError("título", 10, 60)
	}
	if len(o.About) < 200 || len(o.About) > 5000 {
		fmt.Printf("about: %d", len(o.About))
		return shared.Offer{}, lenError("acerca de", 200, 5000)
	}
	if len(o.Requirements) < 200 || len(o.Requirements) > 5000 {
		return shared.Offer{}, lenError("requisitos", 200, 5000)
	}
	if len(o.Benefits) < 200 || len(o.Benefits) > 5000 {
		return shared.Offer{}, lenError("beneficios", 200, 5000)
	}
	minWage, err := strconv.Atoi(o.MinWage)
	if err != nil {
		return shared.Offer{}, fmt.Errorf("el salario mínimo debe ser un número")
	}
	maxWage, err := strconv.Atoi(o.MaxWage)
	if err != nil {
		return shared.Offer{}, fmt.Errorf("el salario máximo debe ser un número")
	}
	if minWage < 0 || maxWage < 0 {
		return shared.Offer{}, fmt.Errorf("el salario no puede ser negativo")
	}
	return shared.Offer{
		Title:        o.Title,
		About:        o.About,
		Requirements: o.Requirements,
		Benefits:     o.Benefits,
		MinWage:      shared.IntToInt32(minWage),
		MaxWage:      shared.IntToInt32(maxWage),
		CompanyID:    o.CompanyID,
	}, nil
}

func ValidateQuiz(q QuizJ) (shared.Quiz, error) {
	duration, err := strconv.Atoi(q.Duration)
	if err != nil {
		return shared.Quiz{}, fmt.Errorf("la duración de la prueba debe ser un número")
	} else if duration < 15 || duration > 180 {
		return shared.Quiz{}, lenError("duración", 0, 180)
	}
	languages := []int32{}
	if len(q.Languages) < 1 || len(q.Languages) > availableLanguages {
		return shared.Quiz{}, fmt.Errorf("debe haber entre 1 y %d lenguajes", availableLanguages)
	} else {
		for _, l := range q.Languages {
			intLang, err := strconv.Atoi(l)
			if err != nil {
				return shared.Quiz{}, fmt.Errorf("el lenguaje debe ser un número")
			}
			languages = append(languages, shared.IntToInt32(intLang))
		}
	}
	return shared.Quiz{
		Duration:  shared.IntToInt32(duration),
		Languages: languages,
	}, nil
}

func ValidateProblem(p ProblemJ) (shared.Problem, error) {
	if len(p.Title) < 1 || len(p.Title) > 64 {
		return shared.Problem{}, lenError("título", 1, 64)
	}
	if len(p.Description) < 10 || len(p.Description) > 5000 {
		fmt.Println("title", len(p.Description))
		return shared.Problem{}, lenError("descripción", 10, 500)
	}
	timeLimit, err := strconv.Atoi(p.TimeLimit)
	if err != nil {
		return shared.Problem{}, fmt.Errorf("el límite de tiempo debe ser un número")
	} else if timeLimit < 500 || timeLimit > 2000 {
		return shared.Problem{}, fmt.Errorf("el límite de tiempo debe estar entre 500 y 2000 ms")
	}
	tcs := []shared.TestCase{}
	if len(p.TestCases) < 1 || len(p.TestCases) > 10 {
		return shared.Problem{}, fmt.Errorf("debe haber entre 1 y 10 casos de prueba")
	} else {
		for _, tc := range p.TestCases {
			if len(tc.Input) < 1 || len(tc.Input) > 1000 {
				return shared.Problem{}, lenError("entrada", 1, 1000)
			}
			if len(tc.Output) < 1 || len(tc.Output) > 1000 {
				return shared.Problem{}, lenError("salida", 1, 1000)
			}
			tcs = append(tcs, shared.TestCase{Input: tc.Input, Output: tc.Output})
		}
	}
	exs := []shared.Example{}
	if len(p.Examples) < 1 || len(p.Examples) > 10 {
		return shared.Problem{}, fmt.Errorf("debe haber entre 1 y 10 ejemplos")
	} else {
		for _, ex := range p.Examples {
			if len(ex.Input) < 1 || len(ex.Input) > 1000 {
				return shared.Problem{}, lenError("entrada", 1, 1000)
			}
			if len(ex.Output) < 1 || len(ex.Output) > 1000 {
				return shared.Problem{}, lenError("salida", 1, 1000)
			}
			exs = append(exs, shared.Example{Input: ex.Input, Output: ex.Output})
		}
	}
	return shared.Problem{
		Title:       p.Title,
		Description: p.Description,
		TimeLimit:   shared.IntToInt32(timeLimit),
		MemoryLimit: memoryLimit,
		Examples:    exs,
		TestCases:   tcs,
	}, nil
}

func GetOfferRegInput(r *http.Request) (OfferRegInput, error) {
	jsonInput, err := shared.Decode[OfferRegJ](r)
	if err != nil {
		return OfferRegInput{}, err
	}
	input := OfferRegInput{}
	offer, err := ValidateOffer(jsonInput.Offer)
	if err != nil {
		return OfferRegInput{}, err
	}
	quiz, err := ValidateQuiz(jsonInput.Quiz)
	if err != nil {
		return OfferRegInput{}, err
	}
	problems := []shared.Problem{}
	for _, p := range jsonInput.Problems {
		problem, err := ValidateProblem(p)
		if err != nil {
			return OfferRegInput{}, err
		}
		problems = append(problems, problem)
	}
	input.Offer = offer
	input.Quiz = quiz
	input.Problems = problems
	return input, nil
}
