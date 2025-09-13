-- name: CreateResult :one
INSERT INTO results (user_id, page_url, link_url, status)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: GetResultByID :one
SELECT * FROM results
WHERE id = $1;

-- name: ListResultsByUser :many
SELECT * FROM results
WHERE user_id = $1
ORDER BY checked_at DESC
    LIMIT $2 OFFSET $3;

-- name: DeleteResultsByUser :exec
DELETE FROM results
WHERE user_id = $1;
