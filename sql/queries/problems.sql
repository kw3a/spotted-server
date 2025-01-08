
-- name: SelectProblem :one
SELECT problem.*
FROM problem
WHERE problem.id = ?;
--

-- name: SelectProblemIDs :many
SELECT problem.id
FROM problem
INNER JOIN quiz ON problem.quiz_id = quiz.id
WHERE quiz.id = ?;

-- name: InsertProblem :exec
INSERT INTO problem
(id, quiz_id, title, description, memory_limit, time_limit)
VALUES (?, ?, ?, ?, ?, ?);
