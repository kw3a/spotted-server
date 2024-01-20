-- +goose Up
CREATE TABLE language_problem (
  id CHAR(36) PRIMARY KEY,
  language_id INTEGER NOT NULL,
  FOREIGN KEY (language_id) REFERENCES language(id) ON DELETE CASCADE,
  problem_id CHAR(36) NOT NULL,
  FOREIGN KEY (problem_id) REFERENCES problem(id) ON DELETE CASCADE,
  UNIQUE (problem_id, language_id)
);

-- +goose Down
DROP TABLE language_problem;