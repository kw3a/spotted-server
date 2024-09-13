-- name: GetQuizzes :many
SELECT * 
FROM quiz
ORDER BY quiz.created_at DESC
LIMIT 10;

-- name: GetQuiz :one
SELECT *
FROM quiz
WHERE quiz.id = ?;
