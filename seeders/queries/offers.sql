-- name: SeedOffer :exec
INSERT INTO offer
(id, status, title, about, requirements, benefits, min_wage, max_wage, company_id) VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteOffers :exec
DELETE FROM offer
WHERE id IN (sqlc.slice('ids'));
