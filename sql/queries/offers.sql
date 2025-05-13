-- name: GetOffer :one
SELECT offer.*, company.name as company_name
FROM offer
JOIN company ON offer.company_id = company.id
WHERE offer.id = ?
LIMIT 1;

-- name: GetOfferByUser :one
SELECT offer.*, company.name as company_name
FROM offer
JOIN company ON offer.company_id = company.id
JOIN user ON company.user_id = user.id
WHERE offer.id = ? AND user.id = ?
LIMIT 1;

-- name: GetOffers :many
SELECT offer.*, company.name as company_name, company.image_url as company_image_url
FROM offer
JOIN company ON offer.company_id= company.id
WHERE offer.status = 1
ORDER BY offer.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetOffersByQuery :many
SELECT offer.*, company.name as company_name, company.image_url as company_image_url
FROM offer
JOIN company ON offer.company_id = company.id
WHERE offer.title LIKE CONCAT('%', ?, '%') AND offer.status = 1
ORDER BY offer.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetOffersByUser :many
SELECT offer.*, company.name as company_name, company.image_url as company_image_url
FROM offer
JOIN company ON offer.company_id = company.id
JOIN user ON company.user_id = user.id
WHERE user.id = ? 
ORDER BY offer.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetOffersByCompany :many
SELECT offer.*, company.name as company_name, company.image_url as company_image_url
FROM offer
JOIN company ON offer.company_id = company.id
WHERE company.id = ? 
ORDER BY offer.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetParticipatedOffers :many
SELECT offer.*, company.name as company_name, company.image_url as company_image_url
FROM offer
JOIN company ON offer.company_id = company.id
JOIN quiz ON offer.id = quiz.offer_id
JOIN participation ON quiz.id = participation.quiz_id
WHERE participation.user_id = ?
ORDER BY participation.expires_at DESC
LIMIT ? OFFSET ?;

-- name: GetOfferByQuiz :one
SELECT offer.*
FROM offer
JOIN quiz ON offer.id = quiz.offer_id
WHERE quiz.id = ?
LIMIT 1;

-- name: InsertOffer :exec
INSERT INTO offer
(id, title, about, requirements, benefits, min_wage, max_wage, company_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ArchiveOffer :exec
UPDATE offer
JOIN company ON offer.company_id = company.id
SET offer.status = -1
WHERE offer.id = ? AND company.user_id = ?;
