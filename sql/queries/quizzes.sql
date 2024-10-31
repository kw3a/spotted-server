-- name: GetQuizzes :many
SELECT quiz.*, user.name as author
FROM quiz
JOIN user ON quiz.user_id = user.id
ORDER BY quiz.created_at DESC
LIMIT 10;

-- name: GetQuiz :one
SELECT quiz.*, user.name as author, GROUP_CONCAT(language.display_name) AS languages 
FROM quiz
JOIN user ON quiz.user_id = user.id
LEFT JOIN language_quiz ON quiz.id = language_quiz.quiz_id
LEFT JOIN language ON language_quiz.language_id = language.id
WHERE quiz.id = ?
GROUP BY quiz.id, user.name;
