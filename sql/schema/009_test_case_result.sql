-- +goose Up

CREATE TABLE test_case_result (
  id CHAR(36) UNIQUE DEFAULT NULL,
  status VARCHAR(64) NOT NULL,
  time DECIMAL(6,3) NOT NULL,
  memory INTEGER NOT NULL,
  test_case_id CHAR(36) NOT NULL,
  FOREIGN KEY (test_case_id) REFERENCES test_case(id) ON DELETE CASCADE,
  submission_id CHAR(36) NOT NULL,
  FOREIGN KEY (submission_id) REFERENCES submission(id) ON DELETE CASCADE,
  UNIQUE (submission_id, test_case_id)
);

-- +goose Down
DROP TABLE test_case_result;