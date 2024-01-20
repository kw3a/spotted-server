-- +goose Up
CREATE TABLE participation (
  id CHAR(36) PRIMARY KEY,
  date TIMESTAMP(0) NOT NULL,
  user_id CHAR(36) NOT NULL,
  FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
  quiz_id CHAR(36) NOT NULL,
  FOREIGN KEY (quiz_id) REFERENCES quiz(id) ON DELETE CASCADE,
  UNIQUE (user_id, quiz_id)
);

-- +goose Down
DROP TABLE participation;