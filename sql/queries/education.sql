-- name: SelectEducation :many
SELECT education.*
FROM education
JOIN user ON education.user_id = user.id
WHERE user.id = ?
ORDER BY education.start_date DESC
LIMIT 10;
-- name: InsertEducation :exec
INSERT INTO education
(id, user_id, institution, degree, start_date, end_date)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteEducation :exec
DELETE FROM education
WHERE id = ? AND user_id = ?;
