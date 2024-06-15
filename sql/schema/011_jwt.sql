-- +goose Up
CREATE TABLE jwt(
  refresh_token VARCHAR(256) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE jwt;