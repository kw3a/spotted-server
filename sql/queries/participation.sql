-- name: ParticipationStatus :one
SELECT participation.id, participation.expires_at
FROM participation
WHERE participation.user_id = ? AND participation.quiz_id = ?;

-- name: Participate :exec
INSERT INTO participation (id, user_id, quiz_id)
VALUES (?, ?, ?);
