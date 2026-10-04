-- name: CreateSubscription :one
INSERT INTO subscriptions (event_type, endpoint_url) VALUES ($1, $2) RETURNING *;

-- name: GetUrlsByEventType :many
SELECT endpoint_url FROM subscriptions WHERE event_type = $1;