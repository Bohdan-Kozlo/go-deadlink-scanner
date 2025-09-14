-- name: CreateSession :one
INSERT INTO sessions (user_id, session_token, expires_at)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE session_token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_token = $1;

-- name: DeleteSessionsByUser :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: GetActiveSessionsByUser :one
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > NOW();
