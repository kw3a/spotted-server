-- name: SeedQuiz :exec
INSERT INTO quiz 
(id, duration, offer_id) VALUES
(?, ?, ?);

-- name: DeleteQuizes :exec
DELETE FROM quiz
WHERE id IN (sqlc.slice('ids'));
