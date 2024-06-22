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
)

//go:embed "static"
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
	//registerRoutes(apiRouter, &apiCfg)
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
	devMiddleware := app.AuthService.DevMiddleware
	fileServer := http.FileServer(http.FS(Files))
	r.Get("/", app.JobOffersHandler())
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	r.Handle("/static/*", fileServer)
	r.Get("/login", app.LoginPageHandler())
	r.Post("/login", app.AuthService.LoginHandler())
	r.Get("/languages/{quizID}", app.LanguagesHandler())
	r.Get("/problems", app.ProblemsHandler())
	r.Get("/examples", app.ExamplesHandler())
	r.Get("/quizzes/{quizID}", devMiddleware(app.QuizPageHandler()))
	r.Get("/source", devMiddleware(app.SourceHandler()))
  r.Get("/score", devMiddleware(app.ScoreHandler()))

  r.Post("/submissions", devMiddleware(app.RunHandler()))
  r.HandleFunc("/results/{submissionID}", devMiddleware(app.ResultsHandler()))
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
