-- name: SeedTestCase :exec
INSERT INTO test_case
(id, input, output, problem_id) VALUES
(?, ?, ?, ?);

-- name: DeleteTestCases :exec
DELETE FROM test_case
WHERE id IN (sqlc.slice('ids'));