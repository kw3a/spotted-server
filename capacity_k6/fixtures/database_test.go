//go:build capacity_internal_sql

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

// Solo usa el servidor efimero indicado explicitamente, nunca .env.
func TestDatabaseSeedAndCleanup(t *testing.T) {
	source := os.Getenv("CAPACITY_TEST_DSN")
	if source == "" {
		t.Skip("CAPACITY_TEST_DSN required for isolated MySQL integration")
	}
	cfg, err := mysql.ParseDSN(source)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBName != "" {
		t.Fatal("integration DSN must not select an existing database")
	}
	cfg.MultiStatements = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("capacity_k6_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP DATABASE " + name)
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, directory := range []string{"schema", "triggers"} {
		files, err := filepath.Glob(filepath.Join("..", "..", "sql", directory, "*.sql"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			up := strings.Split(string(data), "-- +goose Down")[0]
			if _, err = db.Exec(up); err != nil {
				t.Fatalf("%s: %v", file, err)
			}
		}
	}
	ctx := context.Background()
	m, err := seedDataset(ctx, db, false)
	if err != nil {
		t.Fatal(err)
	}
	// Repetir seed conserva exactamente el mismo volumen y relaciones.
	if _, err = seedDataset(ctx, db, false); err != nil {
		t.Fatal(err)
	}
	assertCount := func(table string, want int) {
		t.Helper()
		var got int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s: got %d want %d", table, got, want)
		}
	}
	for table, count := range map[string]int{"user": 10, "company": 10, "offer": 10, "quiz": 10, "problem": 30, "example": 60, "test_case": 150, "language_quiz": 20} {
		assertCount(table, count)
	}
	// Una version anterior requiere actualizacion explicita y transaccional.
	if _, err = db.Exec("UPDATE problem SET title='Suma de enteros',description='legacy' WHERE id=?", m.Offers[0].Problems[0].ProblemID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE test_case SET input='0\n',output='0\n' WHERE id=?", fixtureID("case/01/0/0")); err != nil {
		t.Fatal(err)
	}
	if _, err = seedDataset(ctx, db, false); err == nil {
		t.Fatal("seed silently accepted stale fixtures")
	}
	if _, err = seedDataset(ctx, db, true); err != nil {
		t.Fatal("refresh failed:", err)
	}
	if _, err = seedDataset(ctx, db, true); err != nil {
		t.Fatal("refresh is not idempotent:", err)
	}
	assertCount("problem", 30)
	assertCount("test_case", 150)
	if err = validateDatabase(ctx, db, m); err != nil {
		t.Fatal(err)
	}
	run := "0123456789abcdef"
	other := "fedcba9876543210"
	user := fixtureID("integration/run")
	unrelated := fixtureID("integration/unrelated")
	collision := fixtureID("integration/collision")
	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, u := range []struct{ id, run, nick, name string }{
		{user, run, "ck_" + run + "_1_0", runName(run)},
		{unrelated, other, "ck_" + other + "_1_0", runName(other)},
		{collision, run, "ck_" + run + "_999_0", "unrelated human account"},
	} {
		mustExec("INSERT INTO user (id,nick,name,password) VALUES (?,?,?,?)", u.id, u.nick, u.name, "unused")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: u.id})
		raw, err := token.SignedString([]byte("integration-test"))
		if err != nil {
			t.Fatal(err)
		}
		mustExec("INSERT INTO jwt (refresh_token,created_at) VALUES (?,NOW())", raw)
	}
	mustExec("INSERT INTO jwt (refresh_token,created_at) VALUES ('invalid-other-token',NOW())")
	part := fixtureID("integration/participation")
	submission := fixtureID("integration/submission")
	mustExec("INSERT INTO participation (id,user_id,quiz_id) VALUES (?,?,?)", part, user, m.Offers[0].QuizID)
	if _, err = seedDataset(ctx, db, true); err == nil {
		t.Fatal("refreshed problems with an existing participation")
	}
	if err = validateDatabase(ctx, db, m); err != nil {
		t.Fatal("failed refresh modified fixtures:", err)
	}
	mustExec("INSERT INTO submission (id,src,problem_id,participation_id,language_id) VALUES (?,?,?,?,54)", submission, "int main() { return 0; }", m.Offers[0].Problems[0].ProblemID, part)
	assertCount("test_case_result", 5)
	mustExec("INSERT INTO stroke_window (id,participation_id,stroke_amount,ud_mean,ud_std_dev,du1_mean,du1_std_dev,du2_mean,du2_std_dev,dd_mean,dd_std_dev,uu_mean,uu_std_dev) VALUES (?,?,20,90,15,120,20,210,30,130,20,140,20)", fixtureID("integration/strokes"), part)
	mustExec("INSERT INTO skill (id,name,user_id) VALUES (?,?,?)", fixtureID("integration/skill"), "Go", user)
	mustExec("INSERT INTO link (id,name,url,user_id) VALUES (?,?,?,?)", fixtureID("integration/link"), "profile", "https://example.com", user)
	mustExec("INSERT INTO education (id,institution,degree,start_date,user_id) VALUES (?,?,?,'2020-01-01',?)", fixtureID("integration/education"), "University", "CS", user)
	mustExec("INSERT INTO experience (id,company,title,start_date,user_id) VALUES (?,?,?,'2020-01-01',?)", fixtureID("integration/experience"), "Test", "Developer", user)
	mustExec("INSERT INTO company (id,name,description,website,image_url,user_id) VALUES (?,?,?,'','',?)", fixtureID("integration/company"), "runtime company", "runtime company description", user)
	deleted, err := cleanup(ctx, db, run)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("deleted %d users", deleted)
	}
	assertCount("user", 12)
	assertCount("company", 10)
	assertCount("jwt", 3)
	for _, table := range []string{"participation", "submission", "test_case_result", "stroke_window", "link", "skill", "education", "experience"} {
		assertCount(table, 0)
	}
	if err = validateDatabase(ctx, db, m); err != nil {
		t.Fatal("fixtures changed:", err)
	}
	deleted, err = cleanup(ctx, db, run)
	if err != nil || deleted != 0 {
		t.Fatalf("cleanup not idempotent: %d %v", deleted, err)
	}

}
