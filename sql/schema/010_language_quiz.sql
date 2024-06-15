-- +goose Up
CREATE TABLE language_quiz (
  id CHAR(36) PRIMARY KEY,
  language_id INTEGER NOT NULL,
  FOREIGN KEY (language_id) REFERENCES language(id) ON DELETE CASCADE,
  quiz_id CHAR(36) NOT NULL,
  FOREIGN KEY (quiz_id) REFERENCES quiz(id) ON DELETE CASCADE,
  UNIQUE (quiz_id, language_id)
);

-- +goose Down
DROP TABLE language_quiz;