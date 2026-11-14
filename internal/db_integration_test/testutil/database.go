// Package testutil provides isolated fixtures for server–MySQL integration tests.
package testutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/kw3a/spotted-server/internal/server/storage"
	"github.com/stretchr/testify/require"
)

type Fixture struct {
	DB                                            *sql.DB
	Storage                                       *storage.MysqlStorage
	Ctx                                           context.Context
	UserID, CompanyID, OfferID, QuizID, ProblemID string
	CaseIDs                                       []string
}

func New(t *testing.T) *Fixture {
	t.Helper()
	source := os.Getenv("DB_INTEGRATION_DSN")
	if source == "" {
		t.Skip("DB_INTEGRATION_DSN required; use internal/db_integration_test/run.sh")
	}
	cfg, err := mysql.ParseDSN(source)
	require.NoError(t, err)
	require.Empty(t, cfg.DBName, "DSN must not select an existing database")
	cfg.ParseTime = true
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 10 * time.Second
	cfg.WriteTimeout = 10 * time.Second
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	require.NoError(t, admin.PingContext(ctx))
	name := "spotted_integration_" + uuid.New().String()[:8]
	_, err = admin.ExecContext(ctx, "CREATE DATABASE `"+name+"`")
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, err := admin.ExecContext(cleanupCtx, "DROP DATABASE `"+name+"`")
		require.NoError(t, err)
	})
	cfg.DBName = name
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "..", "..")
	for _, directory := range []string{"schema", "triggers"} {
		command := exec.CommandContext(ctx, "goose", "-dir", filepath.Join(root, "sql", directory), "mysql", cfg.FormatDSN(), "up")
		// Goose failures can print their DSN; only report migration output without credentials.
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("migrate %s: %v (output length %d)", directory, err, len(output))
		}
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	s, err := storage.NewMysqlStorage(cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, s.Close()) })
	f := &Fixture{DB: db, Storage: s, Ctx: ctx, UserID: uuid.NewString(), CompanyID: uuid.NewString(), OfferID: uuid.NewString(), QuizID: uuid.NewString(), ProblemID: uuid.NewString()}
	f.Exec(t, "INSERT INTO language (id,name,display_name) VALUES (54,'cpp','C++'),(62,'java','Java')")
	f.AddUser(t, f.UserID)
	f.Exec(t, "INSERT INTO company (id,user_id,name,description,website,image_url) VALUES (?,?,'Company','Description','https://example.com','')", f.CompanyID, f.UserID)
	f.Exec(t, "INSERT INTO offer (id,title,about,requirements,benefits,min_wage,max_wage,company_id) VALUES (?,'Offer','','','',100,200,?)", f.OfferID, f.CompanyID)
	f.Exec(t, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,30,?)", f.QuizID, f.OfferID)
	f.AddProblem(t, f.ProblemID, f.QuizID)
	for _, input := range []string{"1", "2"} {
		id := uuid.NewString()
		f.CaseIDs = append(f.CaseIDs, id)
		f.Exec(t, "INSERT INTO test_case (id,problem_id,input,output) VALUES (?,?,?,?)", id, f.ProblemID, input, input)
	}
	return f
}

func (f *Fixture) Exec(t *testing.T, query string, args ...interface{}) {
	t.Helper()
	_, err := f.DB.ExecContext(f.Ctx, query, args...)
	require.NoError(t, err)
}
func (f *Fixture) Count(t *testing.T, query string, args ...interface{}) int {
	t.Helper()
	var n int
	require.NoError(t, f.DB.QueryRowContext(f.Ctx, query, args...).Scan(&n))
	return n
}
func (f *Fixture) AddUser(t *testing.T, id string) {
	t.Helper()
	f.Exec(t, "INSERT INTO user (id,nick,name,password) VALUES (?,?, 'Applicant','unused')", id, id[:32])
}
func (f *Fixture) AddProblem(t *testing.T, id, quiz string) {
	t.Helper()
	f.Exec(t, "INSERT INTO problem (id,quiz_id,title,description,memory_limit,time_limit) VALUES (?,?,'Problem','Description',262144,1500)", id, quiz)
}
func (f *Fixture) Participate(t *testing.T, user string) string {
	t.Helper()
	require.NoError(t, f.Storage.Participate(f.Ctx, user, f.QuizID))
	p, err := f.Storage.ParticipationStatus(f.Ctx, user, f.QuizID)
	require.NoError(t, err)
	return p.ID
}
func (f *Fixture) Submit(t *testing.T, participation, problem string, language int32) string {
	t.Helper()
	id := uuid.NewString()
	require.NoError(t, f.Storage.CreateSubmission(f.Ctx, id, participation, problem, "source "+id, language))
	return id
}
func Callback(t *testing.T, token, status string) shared.CallbackJsonInput {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{"token": token, "stdout": "output\n", "time": "0.123", "memory": 1024, "status": map[string]interface{}{"description": status}})
	require.NoError(t, err)
	var input shared.CallbackJsonInput
	require.NoError(t, json.Unmarshal(raw, &input))
	return input
}
