-- +goose Up

CREATE TABLE test_case_result (
  id CHAR(36) PRIMARY KEY,
  status ENUM (
  'Accepted', 
  'Wrong Answer', 
  'Time limit', 
  'Memory limit', 
  'Compile error'
) NOT NULL,
  metrics INTEGER NOT NULL,
  output VARCHAR(255) NOT NULL,
  judge_token VARCHAR(40) NOT NULL,
  test_case_id CHAR(36) NOT NULL,
  FOREIGN KEY (test_case_id) REFERENCES test_case(id) ON DELETE CASCADE,
  submission_id CHAR(36) NOT NULL,
  FOREIGN KEY (submission_id) REFERENCES submission(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE test_case_result;