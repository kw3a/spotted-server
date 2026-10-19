package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var namespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("spotted-server/capacity-k6/fixtures/v1"))

type Language struct {
	ID                int
	Name, DisplayName string
}

var languages = []Language{{54, "cpp", "C++ (GCC 9.2.0)"}, {62, "java", "Java (OpenJDK 13.0.1)"}}

// Alternar por quiz/problema da 15 problemas por lenguaje y ambos en cada quiz.
func languageFor(quizIndex, problemIndex int) string {
	return strconv.Itoa(languages[(quizIndex+problemIndex)%len(languages)].ID)
}

func fixtureID(name string) string { return uuid.NewSHA1(namespace, []byte(name)).String() }

func (a Algorithm) solutions(languageID string) []Solution {
	programs := programsFor(a.Key, languageID)
	return []Solution{
		{Kind: "valid", Source: programs.Source, Expected: Verdict{Accepted: 5}},
		{Kind: "partial", Source: programs.PartialSource, Expected: Verdict{Accepted: a.PartialAccepted, WrongAnswer: 5 - a.PartialAccepted}},
		{Kind: "syntax_error", Source: programs.SyntaxErrorSource, Expected: Verdict{Errors: 5}},
	}
}

func manifest() Manifest {
	m := Manifest{Version: datasetVersion}
	corpus := algorithms()
	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("%02d", i)
		o := Offer{OfferID: fixtureID("offer/" + key), CompanyID: fixtureID("company/" + key), QuizID: fixtureID("quiz/" + key)}
		for p, a := range corpus {
			o.Problems = append(o.Problems, Problem{ProblemID: fixtureID(fmt.Sprintf("problem/%s/%d", key, p)), Title: a.Title, LanguageID: languageFor(i-1, p), Source: programsFor(a.Key, languageFor(i-1, p)).Source, TestCaseCount: 5, Solutions: a.solutions(languageFor(i-1, p))})
		}
		m.Offers = append(m.Offers, o)
	}
	return m
}

func insert(ctx context.Context, tx *sql.Tx, query string, args ...any) error {
	_, err := tx.ExecContext(ctx, query+" ON DUPLICATE KEY UPDATE id=id", args...)
	return err
}

func seedDataset(ctx context.Context, db *sql.DB, refresh bool, reporters ...func(string, ...any)) (Manifest, error) {
	progress := func(string, ...any) {}
	if len(reporters) > 0 && reporters[0] != nil {
		progress = reporters[0]
	}
	m := manifest()
	progress("Abriendo transaccion e insertando lenguajes...")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return m, err
	}
	defer tx.Rollback()
	for _, language := range languages {
		if err = insert(ctx, tx, "INSERT INTO language (id,name,display_name) VALUES (?,?,?)", language.ID, language.Name, language.DisplayName); err != nil {
			return m, err
		}
	}
	pad := strings.Repeat("Datos de precondicion para pruebas de capacidad. ", 8)
	corpus := algorithms()
	for i, o := range m.Offers {
		progress("Fixture %d/10: usuario, compania, oferta, quiz y sus 3 problemas...", i+1)
		key := fmt.Sprintf("%02d", i+1)
		uid := fixtureID("user/" + key)
		raw := make([]byte, 16)
		if _, err = rand.Read(raw); err != nil {
			return m, err
		}
		hashed, e := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(raw)[:30]), bcrypt.DefaultCost)
		if e != nil {
			return m, e
		}
		if err = insert(ctx, tx, "INSERT INTO user (id,nick,name,password) VALUES (?,?,?,?)", uid, "capacity_base_"+key, "Capacity fixture owner "+key, string(hashed)); err != nil {
			return m, err
		}
		if err = insert(ctx, tx, "INSERT INTO company (id,name,description,website,image_url,user_id) VALUES (?,?,?,?,?,?)", o.CompanyID, "Capacity company "+key, "Compania base de pruebas de capacidad.", "", "", uid); err != nil {
			return m, err
		}
		if err = insert(ctx, tx, "INSERT INTO offer (id,title,about,requirements,benefits,status,min_wage,max_wage,company_id) VALUES (?,?,?,?,?,?,?,?,?)", o.OfferID, "Oferta de algoritmos "+key, pad, pad, pad, 1, 1000, 3000, o.CompanyID); err != nil {
			return m, err
		}
		if err = insert(ctx, tx, "INSERT INTO quiz (id,duration,offer_id) VALUES (?,?,?)", o.QuizID, 180, o.OfferID); err != nil {
			return m, err
		}
		for _, language := range languages {
			languageKey := "language-quiz/" + key + "/" + strconv.Itoa(language.ID)
			if err = insert(ctx, tx, "INSERT INTO language_quiz (id,language_id,quiz_id) VALUES (?,?,?)", fixtureID(languageKey), language.ID, o.QuizID); err != nil {
				return m, err
			}
		}
		if refresh {
			// Bloquear el quiz antes de comprobar participaciones y modificar sus casos.
			var lockedID string
			if err = tx.QueryRowContext(ctx, "SELECT id FROM quiz WHERE id=? FOR UPDATE", o.QuizID).Scan(&lockedID); err != nil {
				return m, err
			}
			// Lectura con bloqueo: observar participaciones que hayan confirmado
			// mientras esperabamos el quiz, aun bajo REPEATABLE READ.
			var participationID string
			err = tx.QueryRowContext(ctx, "SELECT id FROM participation WHERE quiz_id=? LIMIT 1 FOR UPDATE", o.QuizID).Scan(&participationID)
			if err != nil && err != sql.ErrNoRows {
				return m, err
			}
			if err == nil {
				return m, fmt.Errorf("no se pueden actualizar problemas con participaciones existentes; limpiar los runs antes de seed --refresh-problems")
			}
		}
		for p, a := range corpus {
			pid := o.Problems[p].ProblemID
			if err = insert(ctx, tx, "INSERT INTO problem (id,title,description,memory_limit,time_limit,quiz_id) VALUES (?,?,?,?,?,?)", pid, a.Title, a.Description, 262144, 1000, o.QuizID); err != nil {
				return m, err
			}
			if refresh {
				if _, err = tx.ExecContext(ctx, "UPDATE problem SET title=?,description=?,memory_limit=262144,time_limit=1000 WHERE id=? AND quiz_id=?", a.Title, a.Description, pid, o.QuizID); err != nil {
					return m, err
				}
			}
			for j, ex := range a.Examples {
				if err = insert(ctx, tx, "INSERT INTO example (id,input,output,problem_id) VALUES (?,?,?,?)", fixtureID(fmt.Sprintf("example/%s/%d/%d", key, p, j)), ex.Input, ex.Output, pid); err != nil {
					return m, err
				}
				if refresh {
					if _, err = tx.ExecContext(ctx, "UPDATE example SET input=?,output=? WHERE id=? AND problem_id=?", ex.Input, ex.Output, fixtureID(fmt.Sprintf("example/%s/%d/%d", key, p, j)), pid); err != nil {
						return m, err
					}
				}
			}
			for j, tc := range a.Cases {
				if err = insert(ctx, tx, "INSERT INTO test_case (id,input,output,problem_id) VALUES (?,?,?,?)", fixtureID(fmt.Sprintf("case/%s/%d/%d", key, p, j)), tc.Input, tc.Output, pid); err != nil {
					return m, err
				}
				if refresh {
					if _, err = tx.ExecContext(ctx, "UPDATE test_case SET input=?,output=? WHERE id=? AND problem_id=?", tc.Input, tc.Output, fixtureID(fmt.Sprintf("case/%s/%d/%d", key, p, j)), pid); err != nil {
						return m, err
					}
				}
			}
		}
	}
	progress("Validando relaciones, problemas, ejemplos, casos secretos y triggers...")
	if err = validateDatabase(ctx, tx, m, progress); err != nil {
		return m, err
	}
	progress("Confirmando transaccion...")
	return m, tx.Commit()
}
