-- name: GetQuizByOffer :one
SELECT quiz.*
FROM quiz
WHERE quiz.offer_id = ?;

-- name: InsertQuiz :exec
INSERT INTO quiz (id, duration, offer_id)
VALUES (?, ?, ?);
