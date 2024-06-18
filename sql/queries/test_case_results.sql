-- name: CreateTestCaseResult :exec
INSERT INTO test_case_result
(id, status, time, memory, test_case_id, submission_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTestCaseResult :one
SELECT *
FROM test_case_result
WHERE id =?;

-- name: UpdateTestCaseResult :exec
UPDATE test_case_result SET id = ?, status = ?, time = ?, memory = ?
WHERE submission_id = ? and test_case_id = ?;

-- name: GetResults :many
SELECT * 
FROM test_case_result
WHERE id IN (sqlc.slice('ids'));
