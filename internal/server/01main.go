package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

//go:embed static
var Files embed.FS

type EnvVariables struct {
	port         string
	dbURL        string
	jwtSecret    string
	judgeURL     string
	judgeHeaders []codejudge.Judge0Header
	myURL        string
}

func Run() error {
	envVars, err := envVariables()
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	setupCors(r)
	viewRoutes(r, envVars)
	srv := &http.Server{
		Addr:              ":" + envVars.port,
		Handler:           r,
		ReadHeaderTimeout: time.Second * 30,
	}
	fmt.Printf("Serving on port: %s\n", envVars.port)
	return srv.ListenAndServe()
}

func viewRoutes(r *chi.Mux, envVars EnvVariables) {
	app, err := NewApp(envVars)
	if err != nil {
		log.Println(err)
	}
	authNMiddleware := func(next http.Handler) http.Handler {
		return auth.AuthNMiddleware(app.Storage, app.AuthType, next)
	}
	authRMiddleware := func(next http.Handler) http.Handler {
		return auth.AuthRMiddleware("/login", auth.AuthRole, next)
	}
	fileServer := http.FileServer(http.FS(Files))
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	r.Handle("/static/*", fileServer)

	r.NotFound(app.NotFoundHandler())
	r.Post("/register", app.UserHandler())

	r.With(authNMiddleware).Group(func(r chi.Router) {
		r.Get("/register", app.UserPageHandler())
		r.Get("/profile/{userID}", app.ProfilePageHandler())
		r.Get("/register/companies", app.CompanyRegistrationPageHandler())
		r.Post("/register/companies", app.CompanyRegistrationHandler())
		r.Get("/companies", app.CompanyListPageHandler())
		r.Get("/companies/{companyID}", app.CompanyPageHandler())
		r.Get("/register/offers", app.OfferRegistrationPage())
		r.Post("/register/offers", app.OfferRegistration())
		r.Patch("/pictures", app.ProfilePicHandler())
		r.Patch("/email", app.UpdateEmail())
		r.Patch("/cell", app.UpdateCell())
		r.Post("/links", app.LinkRegisterHandler())
		r.Delete("/links/{linkID}", app.LinkDeleteHandler())
		r.Post("/skills", app.SkillRegisterHandler())
		r.Delete("/skills/{skillID}", app.SkillDeleteHandler())
		r.Post("/experiences", app.ExperienceRegisterHandler())
		r.Delete("/experiences/{experienceID}", app.ExperienceDeleteHandler())
		r.Post("/education", app.EducationRegisterHandler())
		r.Delete("/education/{educationID}", app.EducationDeleteHandler())
		r.Patch("/descriptions", app.DescrUpdateHandler())
		r.Get("/login", app.LoginPageHandler())
		r.Post("/login", app.LoginHandler())
		r.Post("/logout", app.LogoutHandler())
		r.Get("/", app.JobOffersHandler())
		r.Get("/preamble/{quizID}", app.PreambleHandler())
		r.Get("/offers/admin", app.OffersAdmin())
		r.Get("/offers/admin/{offerID}", app.OfferAdmin())
		r.Patch("/offers/archive/{offerID}", app.OfferArchive())
	})

	r.With(authNMiddleware).With(authRMiddleware).Group(func(r chi.Router) {
		r.Get("/quizes/{quizID}", app.QuizPageHandler())
		r.Get("/problems", app.ProblemsHandler())
		r.Get("/examples", app.ExamplesHandler())
		r.Get("/source", app.LastSrcHandler())
		r.Get("/score", app.ScoreHandler())
		r.Post("/participate", app.ParticipateHandler())
		r.Post("/end", app.EndHandler())

		r.Post("/submissions", app.RunHandler())
		r.HandleFunc("/results/{submissionID}", app.ResultsHandler())
	})

	r.Put("/api/submissions/{submissionID}/tc/{testCaseID}", app.CallbackHandler())
}

func setupCors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}
