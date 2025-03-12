-- name: GetUserByEmail :one
SELECT id, email, password FROM user WHERE email = ?;

-- name: GetUser :one
SELECT * FROM user WHERE id = ?;

-- name: CreateUser :exec
INSERT INTO user 
(id, name, email, password, description, image_url) VALUES 
(?, ?, ?, ?, ?, ?);

-- name: UpdateImage :exec
UPDATE user SET image_url = ? WHERE id = ?;

-- name: SelectApplicants :many
SELECT user.*
FROM user
JOIN participation ON user.id = participation.user_id
WHERE participation.quiz_id = ?;
