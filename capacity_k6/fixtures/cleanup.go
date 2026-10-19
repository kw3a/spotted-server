package main

import (
	"context"
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
)

// Jwt no tiene FK a user. Se inspecciona solo el subject para seleccionar
// tokens ya almacenados; no se usan estos claims para autenticar solicitudes.
func tokenBelongsTo(token string, users map[string]bool) bool {
	parsed, _, err := jwt.NewParser().ParseUnverified(token, &jwt.RegisteredClaims{})
	if err != nil {
		return false
	}
	sub, err := parsed.Claims.GetSubject()
	return err == nil && users[sub]
}

func cleanup(ctx context.Context, db *sql.DB, runID string) (int64, error) {
	prefix, err := runPrefix(runID)
	if err != nil {
		return 0, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	// Igualdad de prefijo, no LIKE: '_' no se interpreta como wildcard.
	predicate := "LEFT(u.nick,?)=? AND u.name=?"
	args := []any{len(prefix), prefix, runName(runID)}
	rows, err := tx.QueryContext(ctx, "SELECT u.id FROM user u WHERE "+predicate+" FOR UPDATE", args...)
	if err != nil {
		return 0, err
	}
	users := map[string]bool{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		users[id] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	if len(users) == 0 {
		return 0, tx.Commit()
	}
	rows, err = tx.QueryContext(ctx, "SELECT refresh_token FROM jwt")
	if err != nil {
		return 0, err
	}
	var tokens []string
	for rows.Next() {
		var token string
		if err = rows.Scan(&token); err != nil {
			rows.Close()
			return 0, err
		}
		if tokenBelongsTo(token, users) {
			tokens = append(tokens, token)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	for _, token := range tokens {
		if _, err = tx.ExecContext(ctx, "DELETE FROM jwt WHERE refresh_token=?", token); err != nil {
			return 0, err
		}
	}
	// Estas FKs no tienen ON DELETE CASCADE. Los datos nunca se seleccionan
	// por fecha ni por un prefijo comun a todas las ejecuciones.
	for _, table := range []string{"link", "skill", "education", "experience", "company"} {
		if _, err = tx.ExecContext(ctx, "DELETE child FROM "+table+" child JOIN user u ON child.user_id=u.id WHERE "+predicate, args...); err != nil {
			return 0, err
		}
	}
	// Cascada: participation -> submission -> test_case_result, y stroke_window.
	result, err := tx.ExecContext(ctx, "DELETE u FROM user u WHERE "+predicate, args...)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, tx.Commit()
}
