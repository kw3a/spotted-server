-- name: GetProblem :many
SELECT problem.*, language.id as language_id, language.name as language_name, language.version as language_version, "" AS input, "" AS output
FROM problem
JOIN language_problem ON problem.id = language_problem.problem_id
JOIN language ON language_problem.language_id = language.id
WHERE problem.quiz_id = ? and problem.id = ?
UNION
SELECT problem.*, 0 AS language_id, "" AS language_name, 0 AS language_version, example.input, example.output
FROM problem
JOIN example ON problem.id = example.problem_id
WHERE problem.quiz_id = ? and problem.id = ?;
--

-- name: CreateProblem :exec
INSERT INTO problem (id, description, title, memory_limit, time_limit, quiz_id)
VALUES (?,?,?,?,?,?);

-- name: GetProblems :many
SELECT problem.*, language.id as language_id, language.name as language_name, language.version as language_version, "" AS input, "" AS output
FROM problem
JOIN language_problem ON problem.id = language_problem.problem_id
JOIN language ON language_problem.language_id = language.id
WHERE problem.quiz_id = ?
UNION
SELECT problem.*, 0 AS language_id, "" AS language_name, 0 AS language_version, example.input, example.output
FROM problem
JOIN example ON problem.id = example.problem_id
WHERE problem.quiz_id = ?;