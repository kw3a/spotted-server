package server

import (
	"net/http"
	"time"

	"github.com/kw3a/spotted-server/internal/server/quizes"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

func (app *App) CallbackHandler() http.HandlerFunc {
	return quizes.CreateCallbackHandler(
		app.Storage,
		app.Stream,
		shared.Decode[shared.CallbackJsonInput],
		quizes.GetCallbackURLParamsInput,
	)
}

func (app *App) RunHandler() http.HandlerFunc {
	return quizes.CreateRunHandler(
		app.Templ,
		app.Storage,
		app.AuthService,
		app.Stream,
		&app.Judge,
		60*time.Second,
		quizes.GetRunInput,
	)
}

func (app *App) ResultsHandler() http.HandlerFunc {
	return quizes.CreateJudgeResultsHandler(
		app.Stream,
		quizes.GetResultsInput,
	)
}

func (DI *App) QuizPageHandler() http.HandlerFunc {
	return quizes.CreateQuizPageHandler(
		DI.Templ,
		DI.Storage,
		DI.AuthService,
		quizes.GetQuizPageInput,
		quizes.SelectFirstProblem,
		quizes.SelectFirstLanguage,
		quizes.EnumerateProblems,
	)
}

func (app *App) ScoreHandler() http.HandlerFunc {
	return quizes.CreateScoreHandler(
		app.Templ,
		app.Storage,
		app.AuthService,
		quizes.GetScoreInput,
	)
}

func (DI *App) LastSrcHandler() http.HandlerFunc {
	return quizes.CreateSourceHandler(
		DI.Storage,
		DI.AuthService,
		quizes.GetSourceInput,
	)
}

func (app *App) ProblemsHandler() http.HandlerFunc {
	return quizes.CreateProblemHandler(app.Templ, app.Storage, quizes.GetProblemsInput)
}

func (DI *App) ParticipateHandler() http.HandlerFunc {
	return quizes.CreateParticipateHandler(DI.Storage, DI.AuthService, quizes.GetParticipateInput)
}

func (DI *App) EndHandler() http.HandlerFunc {
	return quizes.CreateEndHandler(DI.Storage, DI.AuthService, quizes.GetEndInput)
}

func (app *App) ExamplesHandler() http.HandlerFunc {
	return quizes.CreateExamplesHandler(app.Templ, app.Storage, quizes.GetExamplesInput)
}

func (app *App) KeystrokeWindowHandler() http.HandlerFunc {
	return quizes.CreateKeyStrokeWindowHandler(app.Storage, app.AuthService, quizes.GetStrokeWindowInput)
}

func (app *App) KeystrokeReportHandler() http.HandlerFunc {
	return quizes.CreateKeyStrokeReportHandler(app.Templ, app.Storage, quizes.GetKeyStrokeReportInput)
}
