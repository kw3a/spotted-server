-- +goose Up
CREATE TABLE language (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  version INTEGER NOT NULL
);

-- +goose Down
DROP TABLE language;