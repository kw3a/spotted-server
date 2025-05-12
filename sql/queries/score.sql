-- name: BestSubmission :one
SELECT submission.*, language.display_name as language
FROM submission
JOIN participation ON submission.participation_id = participation.id
JOIN language ON submission.language_id = language.id
WHERE submission.problem_id = ? and participation.user_id = ?
ORDER BY submission.accepted_test_cases DESC, submission.created_at ASC
LIMIT 1;

-- name: BatchBestSubmissionsFromParticipation :many
SELECT *
FROM (
    SELECT 
        s.*, 
        problem.title, 
        language.display_name, 
        COUNT(DISTINCT test_case.id) AS total_test_cases,
        ROW_NUMBER() OVER (
            PARTITION BY s.participation_id, s.problem_id 
            ORDER BY s.accepted_test_cases DESC, s.created_at ASC
        ) AS rk
    FROM submission s
    JOIN language ON s.language_id = language.id
    JOIN problem ON s.problem_id = problem.id
    LEFT JOIN test_case ON problem.id = test_case.problem_id
    WHERE s.participation_id IN (sqlc.slice('participation_ids'))
    GROUP BY s.id  
) ranked
WHERE ranked.rk = 1;

-- name: TotalTestCases :one
SELECT COUNT(test_case.id) as total_test_cases
FROM test_case
WHERE test_case.problem_id = ?;
