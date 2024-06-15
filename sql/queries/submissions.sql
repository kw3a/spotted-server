-- name: CreateSubmission :exec
INSERT INTO submission 
(id, src, problem_id, language_id, participation_id)
SELECT ?, ?, problem.id, ?, participation.id
FROM problem
JOIN quiz ON problem.quiz_id = quiz.id
JOIN participation ON quiz.id = participation.quiz_id
WHERE problem.id = ? and participation.user_id = ? and participation.expires_at < NOW();


-- name: LastSubmission :one
SELECT submission.src
FROM submission
JOIN language ON submission.language_id = language.id
JOIN participation ON submission.participation_id = participation.id
WHERE submission.problem_id = ? and submission.language_id = ? and participation.user_id = ?
ORDER BY submission.created_at DESC
LIMIT 1;
