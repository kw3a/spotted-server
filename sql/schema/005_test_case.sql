-- +goose Up
CREATE TABLE test_case (
  id CHAR(36) PRIMARY KEY,
  input VARCHAR(255) NOT NULL,
  output VARCHAR(255) NOT NULL,
  problem_id CHAR(36) NOT NULL,
  FOREIGN KEY (problem_id) REFERENCES problem(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE test_case;