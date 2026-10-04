-- name: CreateEvent :one
INSERT INTO events (event_type, payload) VALUES ($1, $2) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;