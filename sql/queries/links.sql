-- name: SelectLinks :many
SELECT link.*
FROM link
JOIN user ON link.user_id = user.id
WHERE user.id = ?
ORDER BY link.created_at ASC
LIMIT 10;
-- name: InsertLink :exec
INSERT INTO link
(id, url, name, user_id)
VALUES (?, ?, ?, ?);

-- name: CountLinks :one
SELECT COUNT(*) AS count FROM link WHERE user_id = ?;

-- name: DeleteLink :exec
DELETE FROM link
WHERE id = ? and user_id = ?;
