package main

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// Leer grupos completos evita cientos de viajes de red a la base externa.
// Solo se leen los IDs de precondiciones, sin recorrer los datos de los VUs.
func validateDatabase(ctx context.Context, q queryer, m Manifest, reporters ...func(string, ...any)) error {
	progress := func(string, ...any) {}
	if len(reporters) > 0 && reporters[0] != nil {
		progress = reporters[0]
	}
	expected := manifest()
	if !reflect.DeepEqual(m, expected) {
		return fmt.Errorf("manifest invalido: se requieren los 10 fixtures y soluciones de version %d", datasetVersion)
	}
	quizIDs, offerIDs, problemIDs := []any{}, []any{}, []any{}
	offers := map[string]Offer{}
	problems := map[string]Problem{}
	problemQuiz := map[string]string{}
	problemAlgorithm := map[string]Algorithm{}
	keys := map[string]string{}
	corpus := algorithms()
	for i, o := range expected.Offers {
		offerIDs = append(offerIDs, o.OfferID)
		quizIDs = append(quizIDs, o.QuizID)
		offers[o.OfferID] = o
		keys[o.OfferID] = fmt.Sprintf("%02d", i+1)
		for p, problem := range o.Problems {
			problemIDs = append(problemIDs, problem.ProblemID)
			problems[problem.ProblemID] = problem
			problemQuiz[problem.ProblemID] = o.QuizID
			problemAlgorithm[problem.ProblemID] = corpus[p]
		}
	}
	placeholders := func(ids []any) string { return strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",") }
	read := func(stage, query string, args []any, consume func(*sql.Rows) error) error {
		progress("Validacion: %s...", stage)
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("validacion de %s: %w", stage, err)
		}
		defer rows.Close()
		for rows.Next() {
			if err = consume(rows); err != nil {
				return fmt.Errorf("validacion de %s: %w", stage, err)
			}
		}
		if err = rows.Err(); err != nil {
			return fmt.Errorf("validacion de %s: %w", stage, err)
		}
		return nil
	}
	var triggers int
	if err := read("triggers", "SELECT COUNT(*) FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA=DATABASE() AND TRIGGER_NAME IN ('before_participation_insert','insert_test_case_result_submission','update_submission_test_cases')", nil, func(rows *sql.Rows) error { return rows.Scan(&triggers) }); err != nil {
		return err
	}
	if triggers != 3 {
		return fmt.Errorf("faltan triggers: ejecutar migraciones de esquema y triggers")
	}
	seenOffers := map[string]bool{}
	if err := read("usuarios, companias, ofertas y quizzes", "SELECT o.id,o.status,c.id,u.id,u.nick,q.id,q.duration FROM offer o JOIN company c ON c.id=o.company_id JOIN user u ON u.id=c.user_id JOIN quiz q ON q.offer_id=o.id WHERE o.id IN ("+placeholders(offerIDs)+")", offerIDs, func(rows *sql.Rows) error {
		var id, company, user, nick, quiz string
		var status, duration int
		if err := rows.Scan(&id, &status, &company, &user, &nick, &quiz, &duration); err != nil {
			return err
		}
		o, exists := offers[id]
		key := keys[id]
		if !exists || seenOffers[id] || status != 1 || company != o.CompanyID || user != fixtureID("user/"+key) || nick != "capacity_base_"+key || quiz != o.QuizID || duration != 180 {
			return fmt.Errorf("fixture %s: usuario, compania, oferta o quiz incompatible", key)
		}
		seenOffers[id] = true
		return nil
	}); err != nil {
		return err
	}
	if len(seenOffers) != 10 {
		return fmt.Errorf("faltan usuarios, companias, ofertas o quizzes de precondiciones")
	}
	seenLanguages := map[string]bool{}
	if err := read("lenguajes de los quizzes", "SELECT quiz_id,language_id FROM language_quiz WHERE quiz_id IN ("+placeholders(quizIDs)+") AND language_id IN (54,62)", quizIDs, func(rows *sql.Rows) error {
		var quiz string
		var language int
		if err := rows.Scan(&quiz, &language); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%d", quiz, language)
		if seenLanguages[key] {
			return fmt.Errorf("lenguaje duplicado en quiz %s", quiz)
		}
		seenLanguages[key] = true
		return nil
	}); err != nil {
		return err
	}
	for _, o := range expected.Offers {
		for _, language := range languages {
			if !seenLanguages[fmt.Sprintf("%s/%d", o.QuizID, language.ID)] {
				return fmt.Errorf("fixture %s: %s no habilitado", keys[o.OfferID], language.DisplayName)
			}
		}
	}
	seenProblems := map[string]bool{}
	if err := read("problemas", "SELECT id,quiz_id,title,description,memory_limit,time_limit FROM problem WHERE quiz_id IN ("+placeholders(quizIDs)+")", quizIDs, func(rows *sql.Rows) error {
		var id, quiz, title, description string
		var memory, limit int
		if err := rows.Scan(&id, &quiz, &title, &description, &memory, &limit); err != nil {
			return err
		}
		want, exists := problems[id]
		if !exists || seenProblems[id] || quiz != problemQuiz[id] || title != want.Title || description != problemAlgorithm[id].Description || memory != 262144 || limit != 1000 {
			return fmt.Errorf("problema %s incompatible; actualizar precondiciones con seed --refresh-problems, sin runs activos", id)
		}
		seenProblems[id] = true
		return nil
	}); err != nil {
		return err
	}
	if len(seenProblems) != 30 {
		return fmt.Errorf("faltan problemas de precondiciones; se requieren 3 por quiz")
	}
	for _, table := range []string{"example", "test_case"} {
		type expectedSample struct {
			ProblemID string
			Sample    Sample
		}
		want := map[string]expectedSample{}
		for i, o := range expected.Offers {
			for p, problem := range o.Problems {
				samples, kind := corpus[p].Examples, "example"
				if table == "test_case" {
					samples, kind = corpus[p].Cases, "case"
				}
				for j, sample := range samples {
					want[fixtureID(fmt.Sprintf("%s/%02d/%d/%d", kind, i+1, p, j))] = expectedSample{problem.ProblemID, sample}
				}
			}
		}
		seen := map[string]bool{}
		if err := read(table, "SELECT id,problem_id,input,output FROM "+table+" WHERE problem_id IN ("+placeholders(problemIDs)+")", problemIDs, func(rows *sql.Rows) error {
			var id, problem, input, output string
			if err := rows.Scan(&id, &problem, &input, &output); err != nil {
				return err
			}
			sample, exists := want[id]
			if !exists || seen[id] || problem != sample.ProblemID || input != sample.Sample.Input || output != sample.Sample.Output {
				return fmt.Errorf("%s: muestra %s adicional o modificada", table, id)
			}
			seen[id] = true
			return nil
		}); err != nil {
			return err
		}
		if len(seen) != len(want) {
			return fmt.Errorf("%s: faltan muestras; se requieren %d", table, len(want))
		}
	}
	return nil
}
