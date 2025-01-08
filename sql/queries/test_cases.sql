-- name: GetTestCases :many
SELECT problem.time_limit, problem.memory_limit, test_case.id, test_case.input, test_case.output
FROM problem
JOIN test_case 
ON problem.id = test_case.problem_id
WHERE problem_id = ?;

-- name: InsertTestCase :exec
INSERT INTO test_case
(id, problem_id, input, output)
VALUES (?, ?, ?, ?);
