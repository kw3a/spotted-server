-- name: ParticipationStatus :one
SELECT participation.*
FROM participation
WHERE participation.user_id = ? AND participation.quiz_id = ?;

-- name: Participate :exec
INSERT INTO participation (id, user_id, quiz_id)
VALUES (?, ?, ?);

-- name: EndParticipation :exec
UPDATE participation
SET expires_at = ?
WHERE participation.user_id = ? AND participation.quiz_id = ?;

-- name: SelectApplications :many
SELECT user.*, participation.id as participation_id, participation.created_at as participation_created_at, 
  participation.expires_at as participation_expires_at
FROM user
JOIN participation ON user.id = participation.user_id
WHERE participation.quiz_id = ?;

