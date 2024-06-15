-- name: SeedExample :exec
INSERT INTO example 
(id, input, output, problem_id) VALUES
(?, ?, ?, ?);

-- name: DeleteExamples :exec
DELETE FROM example
WHERE id IN (sqlc.slice('ids'))
