-- name: SeedProblem :exec
INSERT INTO problem 
(id, description, title, memory_limit, time_limit, quiz_id) VALUES
(?, ?, ?, ?, ?, ?);

-- name: DeleteProblems :exec
DELETE FROM problem
WHERE id IN (sqlc.slice('ids'));