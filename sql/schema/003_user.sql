-- +goose Up
CREATE TABLE user (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE user;