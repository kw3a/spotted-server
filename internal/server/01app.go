package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/storage"
)

type App struct {
	Templ       *Templates
	Storage     *storage.MysqlStorage
	AuthService *auth.AuthService
	AuthType    *auth.JWTAuth
	Stream      *codejudge.Stream
	Judge       codejudge.Judge0
	Cld         *cloudinary.Cloudinary
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


func NewApp(envVars EnvVariables) (*App, error) {
	views, err := viewsPath()
	if err != nil {
		return nil, err
	}
	templ := newTemplates(views)
	mysqlStorage, err := storage.NewMysqlStorage(envVars.dbURL)
	if err != nil {
		return nil, err
	}
	cloudinaryService, err := cloudinary.New()
	if err != nil {
		return nil, err
	}
	authType := auth.NewJWTAuth(envVars.jwtSecret)
	authService := &auth.AuthService{}
	stream := codejudge.NewStream()
	callbackPath := "/api/submissions/"
	callbackURL := envVars.myURL + callbackPath
	judge := codejudge.NewJudge0(
		envVars.judgeURL,
		callbackURL,
		envVars.judgeHeaders,
	)
	return &App{
		Templ:       templ,
		Storage:     mysqlStorage,
		AuthService: authService,
		AuthType:    authType,
		Stream:      stream,
		Judge:       judge,
		Cld:         cloudinaryService,
	}, nil
}
