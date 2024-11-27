-- name: SelectSkills :many
SELECT skill.*
FROM skill
JOIN user ON skill.user_id = user.id
WHERE user.id = ?
ORDER BY skill.created_at ASC
LIMIT 10;

-- name: InsertSkill :exec
INSERT INTO skill
(id, user_id, name )
VALUES (?, ?, ?);

-- name: DeleteSkill :exec
DELETE FROM skill
WHERE id = ? AND user_id = ?;
