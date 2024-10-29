-- name: SeedQuiz :exec
INSERT INTO quiz 
(id, title, description, duration, min_wage, max_wage, user_id) VALUES
(?, ?, ?, ?, ?, ?, ?);

-- name: DeleteQuizes :exec
DELETE FROM quiz
WHERE id IN (sqlc.slice('ids'));
