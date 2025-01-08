-- name: SelectLanguages :many
SELECT language.*
FROM language
INNER JOIN language_quiz ON language.id = language_quiz.language_id
WHERE language_quiz.quiz_id = ?;

-- name: AllLanguages :many
SELECT *
FROM language
ORDER BY name;

-- name: InsertLanguageQuiz :exec
INSERT INTO language_quiz
(id, quiz_id, language_id)
VALUES (?, ?, ?);
