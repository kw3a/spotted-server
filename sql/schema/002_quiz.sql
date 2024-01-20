-- +goose Up
CREATE TABLE quiz (
  id CHAR(36) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE quiz;