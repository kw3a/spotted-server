-- name: SeedQuiz :exec
INSERT INTO quiz 
(id, title, description, duration) VALUES
(?, ?, ?, ?);

-- name: DeleteQuizes :exec
DELETE FROM quiz
WHERE id IN (sqlc.slice('ids'));