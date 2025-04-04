-- +goose Up

CREATE TABLE test_case_result (
  id CHAR(36) UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  output TEXT NOT NULL, 
  status VARCHAR(64) NOT NULL DEFAULT "",
  time DECIMAL(6,3) NOT NULL DEFAULT 0,
  memory INTEGER NOT NULL DEFAULT 0,
  test_case_id CHAR(36) NOT NULL,
  FOREIGN KEY (test_case_id) REFERENCES test_case(id) ON DELETE CASCADE,
  submission_id CHAR(36) NOT NULL,
  FOREIGN KEY (submission_id) REFERENCES submission(id) ON DELETE CASCADE,
  PRIMARY KEY (submission_id, test_case_id)
);

-- +goose Down
DROP TABLE test_case_result;
