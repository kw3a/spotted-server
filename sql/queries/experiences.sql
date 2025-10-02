-- name: SelectExperience :many
SELECT experience.*
FROM experience
JOIN user ON experience.user_id = user.id
WHERE user.id = ?
ORDER BY experience.start_date DESC
LIMIT 10;

-- name: InsertExperience :exec
INSERT INTO experience
(id, user_id, title, company, start_date, end_date)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CountExperience :one
SELECT COUNT(*) AS count FROM experience WHERE user_id = ?;

-- name: DeleteExperience :exec
DELETE FROM experience
WHERE id = ? AND user_id = ?;
