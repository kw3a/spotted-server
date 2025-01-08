-- +goose Up
CREATE TABLE submission (
  id CHAR(36) PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  src TEXT NOT NULL,
  accepted_test_cases TINYINT UNSIGNED NOT NULL DEFAULT 0,
  problem_id CHAR(36) NOT NULL,
  FOREIGN KEY (problem_id) REFERENCES problem(id) ON DELETE CASCADE,
  participation_id CHAR(36) NOT NULL,
  FOREIGN KEY (participation_id) REFERENCES participation(id) ON DELETE CASCADE,
  language_id INTEGER NOT NULL,
  FOREIGN KEY (language_id) REFERENCES language(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE submission;
