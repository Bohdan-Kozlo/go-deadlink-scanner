-- +goose Up
CREATE TABLE sessions (
id SERIAL PRIMARY KEY,
user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
session_token TEXT UNIQUE NOT NULL,
expires_at TIMESTAMP NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- +goose Down
DROP TABLE sessions;
