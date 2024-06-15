-- name: GetScores :many
SELECT problem.id, MAX(submission.accepted_test_cases) AS accepted_test_cases
FROM submission
INNER JOIN participation ON submission.participation_id = participation.id
INNER JOIN problem ON submission.problem_id = problem.id
WHERE participation.id = ?
GROUP BY problem.id;