-- name: SeedUser :exec
INSERT INTO user 
(id, name, email, password, description, image_url) VALUES
(?, ?, ?, ?, ?, ?);

-- name: DeleteUsers :exec
DELETE FROM user
WHERE id IN (sqlc.slice('ids'));
