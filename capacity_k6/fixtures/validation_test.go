package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

type validationData struct {
	groups  map[string][][]driver.Value
	queries int
	delay   time.Duration
}

func validationFixtureData() *validationData {
	d := &validationData{groups: map[string][][]driver.Value{"triggers": {{int64(3)}}}}
	corpus := algorithms()
	for i, o := range manifest().Offers {
		key := fmt.Sprintf("%02d", i+1)
		d.groups["offer"] = append(d.groups["offer"], []driver.Value{o.OfferID, int64(1), o.CompanyID, fixtureID("user/" + key), "capacity_base_" + key, o.QuizID, int64(180)})
		for _, l := range languages {
			d.groups["language_quiz"] = append(d.groups["language_quiz"], []driver.Value{o.QuizID, int64(l.ID)})
		}
		for p, problem := range o.Problems {
			d.groups["problem"] = append(d.groups["problem"], []driver.Value{problem.ProblemID, o.QuizID, problem.Title, corpus[p].Description, int64(262144), int64(1000)})
			for _, table := range []string{"example", "test_case"} {
				samples, kind := corpus[p].Examples, "example"
				if table == "test_case" {
					samples, kind = corpus[p].Cases, "case"
				}
				for j, sample := range samples {
					d.groups[table] = append(d.groups[table], []driver.Value{fixtureID(fmt.Sprintf("%s/%s/%d/%d", kind, key, p, j)), problem.ProblemID, sample.Input, sample.Output})
				}
			}
		}
	}
	return d
}

type validationConnector struct{ data *validationData }

func (c validationConnector) Connect(context.Context) (driver.Conn, error) {
	return validationConn(c), nil
}
func (c validationConnector) Driver() driver.Driver { return validationDriver{} }

type validationDriver struct{}

func (validationDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("connector required")
}

type validationConn struct{ data *validationData }

func (validationConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (validationConn) Close() error              { return nil }
func (validationConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected write") }
func (c validationConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.data.queries++
	if c.data.delay > 0 {
		timer := time.NewTimer(c.data.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	group := ""
	if strings.Contains(query, "information_schema.TRIGGERS") {
		group = "triggers"
	}
	for _, table := range []string{"offer", "language_quiz", "problem", "example", "test_case"} {
		if strings.Contains(query, "FROM "+table+" ") {
			group = table
		}
	}
	width, count := map[string]int{"triggers": 1, "offer": 7, "language_quiz": 2, "problem": 6, "example": 4, "test_case": 4}[group], 10
	if group == "triggers" {
		count = 0
	}
	if group == "example" || group == "test_case" {
		count = 30
	}
	if group == "" || len(args) != count {
		return nil, fmt.Errorf("unexpected query/arguments: %s (%d)", query, len(args))
	}
	return &validationRows{data: c.data.groups[group], width: width}, nil
}

type validationRows struct {
	data         [][]driver.Value
	index, width int
}

func (r *validationRows) Columns() []string { return make([]string, r.width) }
func (*validationRows) Close() error        { return nil }
func (r *validationRows) Next(dest []driver.Value) error {
	if r.index == len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.index])
	r.index++
	return nil
}
func validationDB(t *testing.T, data *validationData) *sql.DB {
	t.Helper()
	db := sql.OpenDB(validationConnector{data})
	t.Cleanup(func() { db.Close() })
	return db
}

func TestValidationUsesSixQueriesAndChecksAllFixtures(t *testing.T) {
	data := validationFixtureData()
	var stages []string
	err := validateDatabase(context.Background(), validationDB(t, data), manifest(), func(format string, values ...any) { stages = append(stages, fmt.Sprintf(format, values...)) })
	if err != nil {
		t.Fatal(err)
	}
	if data.queries != 6 || len(stages) != 6 {
		t.Fatalf("expected six queries/progress messages: %d %v", data.queries, stages)
	}
}

func TestValidationRejectsIncompleteOrModifiedFixtures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*validationData)
	}{
		{"missing_trigger", func(d *validationData) { d.groups["triggers"][0][0] = int64(2) }},
		{"missing_offer", func(d *validationData) { d.groups["offer"] = d.groups["offer"][:9] }},
		{"wrong_owner", func(d *validationData) { d.groups["offer"][0][3] = "unrelated" }},
		{"missing_language", func(d *validationData) { d.groups["language_quiz"] = d.groups["language_quiz"][1:] }},
		{"duplicate_language", func(d *validationData) {
			d.groups["language_quiz"] = append(d.groups["language_quiz"], d.groups["language_quiz"][0])
		}},
		{"missing_problem", func(d *validationData) { d.groups["problem"] = d.groups["problem"][1:] }},
		{"extra_problem", func(d *validationData) {
			d.groups["problem"] = append(d.groups["problem"], []driver.Value{"extra", manifest().Offers[0].QuizID, "extra", "extra", int64(262144), int64(1000)})
		}},
		{"modified_problem", func(d *validationData) { d.groups["problem"][0][3] = "legacy" }},
		{"missing_example", func(d *validationData) { d.groups["example"] = d.groups["example"][1:] }},
		{"extra_case", func(d *validationData) {
			d.groups["test_case"] = append(d.groups["test_case"], []driver.Value{"extra", manifest().Offers[0].Problems[0].ProblemID, "x", "x"})
		}},
		{"modified_case", func(d *validationData) { d.groups["test_case"][0][3] = "wrong\n" }},
		{"case_wrong_problem", func(d *validationData) { d.groups["test_case"][0][1] = manifest().Offers[0].Problems[1].ProblemID }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := validationFixtureData()
			tc.mutate(data)
			if err := validateDatabase(context.Background(), validationDB(t, data), manifest()); err == nil {
				t.Fatal("accepted incompatible fixture")
			}
		})
	}
	data := validationFixtureData()
	m := manifest()
	m.Offers[0].Problems[0].Solutions[0].Source = "broken"
	if err := validateDatabase(context.Background(), validationDB(t, data), m); err == nil || data.queries != 0 {
		t.Fatal("invalid manifest reached database")
	}
}
