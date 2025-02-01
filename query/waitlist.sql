-- name: GetAll :many
SELECT * FROM waitlist;

-- name: GetByID :one
SELECT * FROM waitlist WHERE id = ?;

-- name: CreateEntry :execresult
INSERT INTO waitlist (user_id, first_name, last_name, username, bot_username, message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, now(), now());