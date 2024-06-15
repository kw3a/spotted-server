-- name: SaveRefreshJWT :exec
INSERT INTO jwt
(refresh_token, created_at) VALUES 
(?, ?);
--

-- name: VerifyRefreshJWT :one
SELECT COUNT(*) > 0 AS token_exists
FROM jwt
WHERE refresh_token = ?; 
--

-- name: RevokeRefreshJWT :exec
DELETE FROM jwt WHERE refresh_token = ?;
--