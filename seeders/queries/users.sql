-- name: SeedUser :exec
INSERT INTO user 
(id, name, email, password, role) VALUES
(?, ?, ?, ?, ?);

-- name: DeleteUsers :exec
DELETE FROM user
WHERE id IN (sqlc.slice('ids'));