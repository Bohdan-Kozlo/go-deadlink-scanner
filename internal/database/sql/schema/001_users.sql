-- +goose Up
CREATE TABLE users (
id SERIAL PRIMARY KEY,
email VARCHAR(256) UNIQUE NOT NULL,
password VARCHAR(50) NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;