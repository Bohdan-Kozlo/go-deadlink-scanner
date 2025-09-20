-- +goose Up
CREATE TABLE results (
id SERIAL PRIMARY KEY,
user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
page_url TEXT NOT NULL,
link_url TEXT NOT NULL,
status VARCHAR(100) NOT NULL,
checked_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_results_user_id ON results(user_id);


-- +goose Down
DROP TABLE results;
