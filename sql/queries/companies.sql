-- name: InsertCompany :exec
INSERT INTO company
(id, user_id, name, description, website, image_url)
VALUES (?, ?, ?, ?, ?, ?);

-- name: SelectCompany :one
SELECT company.*
FROM company
WHERE company.id = ? AND company.user_id = ?;

-- name: GetCompanies :many
SELECT company.*
FROM company
LIMIT 10;

-- name: GetCompaniesByUser :many
SELECT company.*
FROM company
WHERE company.user_id = ?
LIMIT 10;

-- name: GetCompaniesByQuery :many
SELECT company.*
FROM company
WHERE company.name LIKE CONCAT('%', ?, '%')
LIMIT 10;

-- name: GetCompaniesByUserAndQuery :many
SELECT company.*
FROM company
WHERE company.name LIKE CONCAT('%', ?, '%') AND company.user_id = ?
LIMIT 10;

-- name: GetCompanyByID :one
SELECT company.*
FROM company
WHERE company.id = ?;
