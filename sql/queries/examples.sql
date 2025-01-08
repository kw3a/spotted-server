-- name: SelectExamples :many
SELECT example.input, example.output
FROM example
WHERE example.problem_id = ?;

-- name: InsertExample :exec
INSERT INTO example
(id, problem_id, input, output)
VALUES (?, ?, ?, ?);
