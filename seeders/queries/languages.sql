-- name: SeedLanguage :exec
INSERT INTO language
(id, name, display_name) VALUES
(?, ?, ?);

-- name: DeleteLanguages :exec
DELETE FROM language
WHERE id IN (sqlc.slice('ids'));
