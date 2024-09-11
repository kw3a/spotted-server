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

func NewApp(dbURL, jwtSecret, judgeURL, judgeAuthToken, callbackURL string) (*App, error) {
	views, err := viewsPath()
	if err != nil {
		return nil, err
	}
	templ := newTemplates(views)
	mysqlStorage, err := NewMysqlStorage(dbURL)
	if err != nil {
		return nil, err
	}
	authType := auth.NewJWTAuth(jwtSecret)
	authService := &auth.AuthService{}
	stream := codejudge.NewStream()
	judge := codejudge.NewJudge0(
		judgeURL,
		judgeAuthToken,
		callbackURL,
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
