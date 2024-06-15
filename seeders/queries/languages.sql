-- name: SeedLanguage :exec
INSERT INTO language
(id, name, version) VALUES
(?, ?, ?);

-- name: DeleteLanguages :exec
DELETE FROM language
WHERE id IN (sqlc.slice('ids'));