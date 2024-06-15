package codejudge

/*
import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/auth"
	"github.com/kw3a/spotted-server/internal/server/utils"
)

type CodeJudge struct {
	judge       JudgeService
	storage     CodeJudgeStorage
	authService auth.AuthService
	stream      *Stream
}

var logError = utils.ErrorLog
var logJsonEncode = func(w http.ResponseWriter, status int, data interface{}) {
	logError(utils.Encode(w, status, data))
}

func msgError(msg string) utils.ErrorMesage {
	return utils.ErrorMesage{
		Msg: msg,
	}
}

func NewCodeJudge(queries *database.Queries, db *sql.DB, authService auth.AuthService) *CodeJudge {
	return &CodeJudge{
		judge: &Judge0{
			CallbackURL: os.Getenv("OWN_URL") + "/api/submissions/",
			JudgeURL:    os.Getenv("JUDGE_URL"),
			AuthToken:   os.Getenv("JUDGE_AUTHN_SECRET"),
		},
		storage:     NewCodeJudgeStorage(queries, db),
		authService: authService,
		stream:      NewStream(),
	}
}
*/
