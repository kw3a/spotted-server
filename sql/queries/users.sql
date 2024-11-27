-- name: GetUserByEmail :one
SELECT id, email, password FROM user WHERE email = ?;

-- name: GetUser :one
SELECT * FROM user WHERE id = ?;

-- name: CreateUser :exec
INSERT INTO user 
(id, name, email, password, role, description) VALUES 
(?, ?, ?, ?, ?, ?);

-- name: UpdateImage :exec
UPDATE user SET image_url = ? WHERE id = ?;
