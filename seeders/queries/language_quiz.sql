-- name: SeedLanguageQuiz :exec
INSERT INTO language_quiz
(id, language_id, quiz_id) VALUES
(?, ?, ?);

-- name: DeleteLanguageQuiz :exec
DELETE FROM language_quiz
WHERE id IN (sqlc.slice('ids'));