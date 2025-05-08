-- name: SeedUser :exec
INSERT INTO user 
(id, nick, name, email, password, description, number, image_url) VALUES
(?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteUsers :exec
DELETE FROM user
WHERE id IN (sqlc.slice('ids'));
