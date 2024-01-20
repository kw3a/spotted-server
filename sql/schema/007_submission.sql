-- +goose Up
CREATE TABLE submission (
  id CHAR(36) PRIMARY KEY,
  src TEXT NOT NULL,
  time TIMESTAMP(0) NOT NULL,
  problem_id CHAR(36) NOT NULL,
  FOREIGN KEY (problem_id) REFERENCES problem(id) ON DELETE CASCADE,
  user_id CHAR(36) NOT NULL,
  FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
  language_id INTEGER NOT NULL,
  FOREIGN KEY (language_id) REFERENCES language(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE submission;