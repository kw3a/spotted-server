-- name: CreateSubmission :exec
INSERT INTO submission 
(id, src, language_id, problem_id, participation_id)
VALUES (?, ?, ?, ?, ?);

-- name: LastSubmission :one
SELECT submission.src
FROM submission
JOIN language ON submission.language_id = language.id
JOIN participation ON submission.participation_id = participation.id
WHERE submission.problem_id = ? and submission.language_id = ? and participation.user_id = ?
ORDER BY submission.created_at DESC
LIMIT 1;
