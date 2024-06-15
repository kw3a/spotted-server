package server

import (
	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
)

type App struct {
  Templ *Templates
	Storage     *MysqlStorage
	AuthService *auth.AuthService
	Stream      *codejudge.Stream
	Judge       codejudge.Judge0
}

func NewApp(dbURL, jwtSecret, judgeURL, judgeAuthToken, callbackURL string) (*App, error) {
  templ := newTemplates()
	mysqlStorage, err := NewMysqlStorage(dbURL)
	if err != nil {
		return nil, err
	}
	authService, err := auth.NewAuthService(jwtSecret, dbURL)
	if err != nil {
		return nil, err
	}
	stream := codejudge.NewStream()
	judge := codejudge.NewJudge0(
		judgeURL,
		judgeAuthToken,
		callbackURL,
	)
	return &App{
    Templ: templ,
		Storage:     mysqlStorage,
		AuthService: authService,
		Stream:      stream,
		Judge:       judge,
	}, nil
}
