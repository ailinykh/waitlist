-- name: GetAllEntries :many
SELECT * FROM waitlist;

-- name: GetEntryByID :one
SELECT * FROM waitlist WHERE id = $1;

-- name: CreateEntry :execresult
INSERT INTO waitlist (user_id, first_name, last_name, username, bot_username, message)
VALUES ($1, $2, $3, $4, $5, $6);


-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserByUserID :one
SELECT * FROM users WHERE user_id = $1;

-- name: CreateUser :execresult
INSERT INTO users (user_id, first_name, last_name, username, photo_url, role)
VALUES ($1, $2, $3, $4, $5, 'user');
