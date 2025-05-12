-- name: SelectExamples :many
SELECT example.input, example.output
FROM example
WHERE example.problem_id = ?;

-- name: InsertExample :exec
INSERT INTO example
(id, problem_id, input, output)
VALUES (?, ?, ?, ?);

-- name: BatchExamples :many
SELECT example.*
FROM example
WHERE problem_id IN (sqlc.slice('problem_ids'));

