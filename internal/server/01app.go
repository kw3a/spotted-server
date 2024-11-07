package server

import (
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

type App struct {
	Templ       *Templates
	Storage     *MysqlStorage
	AuthService *auth.AuthService
	AuthType    *auth.JWTAuth
	Stream      *codejudge.Stream
	Judge       codejudge.Judge0
}

func NewApp(envVars EnvVariables) (*App, error) {
	views, err := viewsPath()
	if err != nil {
		return nil, err
	}
	templ := newTemplates(views)
	mysqlStorage, err := NewMysqlStorage(envVars.dbURL)
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
	}, nil
}
