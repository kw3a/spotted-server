-- name: GetUserByNick :one
SELECT id, nick, password FROM user WHERE nick = ?;

-- name: GetUser :one
SELECT * FROM user WHERE id = ?;

-- name: CreateUser :exec
INSERT INTO user 
(id, nick, name, password) VALUES 
(?, ?, ?, ?);

-- name: UpdateImage :exec
UPDATE user SET image_url = ? WHERE id = ?;

-- name: SelectApplicants :many
SELECT user.*
FROM user
JOIN participation ON user.id = participation.user_id
WHERE participation.quiz_id = ?;

-- name: UpdateUserDescription :exec
UPDATE user SET description = ? WHERE id = ?;

-- name: UpdateEmail :exec
UPDATE user SET email = ? WHERE id = ?;

-- name: UpdateCell :exec
UPDATE user SET number = ? WHERE id = ?;
