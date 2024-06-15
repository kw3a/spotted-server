-- +goose Up
CREATE TABLE user (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  role VARCHAR(10) NOT NULL,
  description VARCHAR(500)
);

-- +goose Down
DROP TABLE user;