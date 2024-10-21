-- name: GetQuizzes :many
SELECT quiz.id, quiz.title, user.name as author
FROM quiz
JOIN user ON quiz.user_id = user.id
ORDER BY quiz.created_at DESC
LIMIT 10;

-- name: GetQuiz :one
SELECT quiz.*, user.name as author
FROM quiz
JOIN user ON quiz.user_id = user.id
WHERE quiz.id = ?;
