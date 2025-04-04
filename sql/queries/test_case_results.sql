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
SELECT * 
FROM test_case_result
WHERE id IN (sqlc.slice('ids'));

-- name: GetExecutedTestCases :many
SELECT test_case.id as tc_id, test_case.input as tc_input, test_case.output as tc_output, 
  test_case_result.status as result_status, test_case_result.time as result_time, 
  test_case_result.memory as result_memory, test_case_result.output as result_output
FROM test_case
JOIN test_case_result ON test_case.id = test_case_result.test_case_id 
JOIN submission ON test_case_result.submission_id = submission.id
WHERE test_case.problem_id = ? AND submission.id = ?;
