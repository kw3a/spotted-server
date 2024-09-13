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
