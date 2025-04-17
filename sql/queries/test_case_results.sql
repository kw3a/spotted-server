-- name: CreateTestCaseResult :exec
INSERT INTO test_case_result
(id, status, time, memory, test_case_id, submission_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTestCaseResult :one
SELECT *
FROM test_case_result
WHERE id =?;

-- name: UpdateTestCaseResult :exec
UPDATE test_case_result SET id = ?, status = ?, time = ?, memory = ?, output = ?
WHERE submission_id = ? and test_case_id = ?;

-- name: GetResults :many
SELECT test_case_result.* 
FROM test_case_result
JOIN submission ON test_case_result.submission_id = submission.id
JOIN test_case ON test_case_result.test_case_id = test_case.id
WHERE test_case.problem_id = ? AND submission.id = ?;
