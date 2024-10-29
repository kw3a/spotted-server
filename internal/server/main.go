package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kw3a/spotted-server/internal/auth"
)

//go:embed static
var Files embed.FS

type EnvVariables struct {
	port           string
	dbURL          string
	jwtSecret      string
	judgeURL       string
	judgeAuthToken string
	myURL          string
}

func Run() error {
	envVars, err := envVariables()
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	setupCors(r)
	callbackPath := "/api/submissions/"
	callbackURL := envVars.myURL + callbackPath
	viewRoutes(r, envVars.dbURL, envVars.jwtSecret, envVars.judgeURL, envVars.judgeAuthToken, callbackURL)
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
	judgeAuthToken := os.Getenv("JUDGE_AUTHN_SECRET")
	if judgeAuthToken == "" {
		return EnvVariables{}, fmt.Errorf("JUDGE_AUTHN_SECRET environment variable is not set")
	}
	myURL := os.Getenv("MY_URL")
	if myURL == "" {
		return EnvVariables{}, fmt.Errorf("MY_URL environment variable is not set")
	}
	return EnvVariables{
		port:           port,
		dbURL:          dbURL,
		jwtSecret:      jwtSecret,
		judgeURL:       judgeURL,
		judgeAuthToken: judgeAuthToken,
		myURL:          myURL,
	}, nil

}

func viewRoutes(r *chi.Mux, dbURL, jwtSecret, judgeURL, judgeAuthToken, callbackURL string) {
	app, err := NewApp(dbURL, jwtSecret, judgeURL, judgeAuthToken, callbackURL)
	if err != nil {
		log.Println(err)
	}
	authNMiddleware := func(next http.Handler) http.Handler {
		return auth.AuthNMiddleware(app.Storage, app.AuthType, next)
	}
	authRMiddlewareDev := func(next http.Handler) http.Handler {
		return auth.AuthRMiddleware("/login", "dev", next)
	}
	fileServer := http.FileServer(http.FS(Files))
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	r.Handle("/static/*", fileServer)

	r.NotFound(app.NotFoundHandler())

	r.With(authNMiddleware).Group(func(r chi.Router) {
		r.Get("/login", app.LoginPageHandler())
		r.Post("/login", app.LoginHandler())
		r.Post("/logout", app.LogoutHandler())
		//r.Get("/languages/{quizID}", app.LanguagesHandler())
		r.Get("/", app.JobOffersHandler())
		r.Get("/preamble/{quizID}", app.PreambleHandler())
	})

	r.With(authNMiddleware).With(authRMiddlewareDev).Group(func(r chi.Router) {
		r.Get("/quizzes/{quizID}", app.QuizPageHandler())
		r.Get("/problems", app.ProblemsHandler())
		r.Get("/examples", app.ExamplesHandler())
		r.Get("/source", app.SourceHandler())
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
