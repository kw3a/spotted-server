-- name: CreateSubmission :exec
INSERT INTO submission 
(id, src, time, problem_id, user_id, language_id)
VALUES (?, ?, ?, ?, ?, ?);