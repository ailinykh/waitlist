-- name: GetAllEntries :many
SELECT * FROM waitlist;

-- name: GetEntryByID :one
SELECT * FROM waitlist WHERE id = ?;

-- name: CreateEntry :execresult
INSERT INTO waitlist (user_id, first_name, last_name, username, bot_username, message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, now(), now());


-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserByUserID :one
SELECT * FROM users WHERE user_id = ?;

-- name: CreateUser :execresult
INSERT INTO users (user_id, first_name, last_name, username, photo_url, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, "user", now(), now());
