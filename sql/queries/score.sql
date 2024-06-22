-- name: BestSubmission :one
SELECT submission.*
FROM submission
JOIN participation ON submission.participation_id = participation.id
WHERE submission.problem_id = ? and participation.user_id = ?
ORDER BY submission.accepted_test_cases DESC, submission.created_at ASC
LIMIT 1;


-- name: TotalTestCases :one
SELECT COUNT(test_case.id) as total_test_cases
FROM test_case
WHERE test_case.problem_id = ?;
