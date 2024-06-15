-- name: SelectExamples :many
SELECT example.input, example.output
FROM example
WHERE example.problem_id = ?;