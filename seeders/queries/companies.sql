-- name: SeedCompany :exec
INSERT INTO company
(id, user_id, name, description, website, image_url)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteCompanies :exec
DELETE FROM company
WHERE id IN (sqlc.slice('ids'));
