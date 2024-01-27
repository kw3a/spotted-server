-- name: GetUserByEmail :one
SELECT id, email, password FROM user WHERE email = ?;

-- name: GetUser :one
SELECT * FROM user WHERE id = ?;