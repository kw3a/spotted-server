-- name: SelectLanguages :many
SELECT language.*
FROM language
INNER JOIN language_quiz ON language.id = language_quiz.language_id
INNER JOIN quiz ON language_quiz.quiz_id = quiz.id
WHERE quiz.id = ?;