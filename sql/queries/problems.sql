
-- name: SelectProblem :one
SELECT problem.title, problem.description
FROM problem
WHERE problem.id = ?;
--

-- name: CreateProblem :exec
INSERT INTO problem (id, description, title, memory_limit, time_limit, quiz_id)
VALUES (?,?,?,?,?,?);

-- name: SelectProblemIDs :many
SELECT problem.id
FROM problem
INNER JOIN quiz ON problem.quiz_id = quiz.id
WHERE quiz.id = ?;