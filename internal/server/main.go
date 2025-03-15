package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func envVariables() (EnvVariables, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return EnvVariables{}, fmt.Errorf("PORT environment variable is not set")
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		return EnvVariables{}, fmt.Errorf("PORT environment variable is not a valid integer")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return EnvVariables{}, fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return EnvVariables{}, fmt.Errorf("JWT_SECRET environment variable is not set")
	}
	judgeURL := os.Getenv("JUDGE_URL")
	if judgeURL == "" {
		return EnvVariables{}, fmt.Errorf("JUDGE_URL environment variable is not set")
	}
	judgeHeaders := []codejudge.Judge0Header{}
	if strings.Contains(judgeURL, "rapidapi") {
		rapidAPIKey := os.Getenv("X_RAPID_API_KEY")
		if rapidAPIKey == "" {
			return EnvVariables{}, fmt.Errorf("RAPID_API_KEY environment variable is not set")
		}
		judgeHeaders = append(judgeHeaders, codejudge.Judge0Header{
			Name:  "x-rapidapi-key",
			Value: rapidAPIKey,
		})
		judgeHeaders = append(judgeHeaders, codejudge.Judge0Header{
			Name:  "x-rapidapi-host",
			Value: "judge0.p.rapidapi.com",
		})
	} else {
		judgeAuthToken := os.Getenv("JUDGE_AUTHN_SECRET")
		if judgeAuthToken == "" {
			return EnvVariables{}, fmt.Errorf("JUDGE_AUTHN_SECRET environment variable is not set")
		}
		judgeHeaders = append(judgeHeaders, codejudge.Judge0Header{
			Name:  "X-Auth-Token",
			Value: judgeAuthToken,
		})
	}
	myURL := os.Getenv("MY_URL")
	if myURL == "" {
		return EnvVariables{}, fmt.Errorf("MY_URL environment variable is not set")
	}
	return EnvVariables{
		port:         port,
		dbURL:        dbURL,
		jwtSecret:    jwtSecret,
		judgeURL:     judgeURL,
		judgeHeaders: judgeHeaders,
		myURL:        myURL,
	}, nil

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
		r.Post("/links", app.LinkRegisterHandler())
		r.Delete("/links/{linkID}", app.LinkDeleteHandler())
		r.Post("/skills", app.SkillRegisterHandler())
		r.Delete("/skills/{skillID}", app.SkillDeleteHandler())
		r.Post("/experiences", app.ExperienceRegisterHandler())
		r.Delete("/experiences/{experienceID}", app.ExperienceDeleteHandler())
		r.Post("/education", app.EducationRegisterHandler())
		r.Delete("/education/{educationID}", app.EducationDeleteHandler())
		r.Get("/login", app.LoginPageHandler())
		r.Post("/login", app.LoginHandler())
		r.Post("/logout", app.LogoutHandler())
		//r.Get("/languages/{quizID}", app.LanguagesHandler())
		r.Get("/", app.JobOffersHandler())
		r.Get("/preamble/{quizID}", app.PreambleHandler())
		r.Get("/offers/admin", app.OffersAdmin())
		r.Get("/offers/admin/{offerID}", app.OfferAdmin())
		r.Get("/source/{problemID}/{applicantID}", app.SourceHandler())
		r.Patch("/offers/archive/{offerID}", app.OfferArchive())
	})

	r.With(authNMiddleware).With(authRMiddleware).Group(func(r chi.Router) {
		r.Get("/quizzes/{quizID}", app.QuizPageHandler())
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
