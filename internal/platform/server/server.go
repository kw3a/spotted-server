package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/kw3a/spotted-server/internal/database"
)

type ApiConfig struct {
	DB *database.Queries
}

func Run(port int) error {

	apiCfg := ApiConfig{}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			log.Fatal(err)
			return err
		}
		dbQueries := database.New(db)
		apiCfg.DB = dbQueries
		log.Println("Connected to database!")
	}

	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	registerRoutes(apiRouter, &apiCfg)
	r.Mount("/api", apiRouter)
	strPort := strconv.Itoa(port)
	srv := &http.Server{
		Addr:              ":" + strPort,
		Handler:           r,
		ReadHeaderTimeout: 30 * time.Second,
	}
	fmt.Printf("Serving on port: %s\n", strPort)
	return srv.ListenAndServe()
}

func registerRoutes(r *chi.Mux, apiCfg *ApiConfig) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/quizzes/{quizID}", apiCfg.handlerProblemsGet)
	r.Get("/quizzes/{quizID}/problems/{problemID}", apiCfg.handlerProblemGet)
	r.Post("/problems", apiCfg.handlerProblemCreate)
	r.Get("/quizzes", apiCfg.handlerQuizzesGet)
}
