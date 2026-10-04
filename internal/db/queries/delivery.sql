-- name: CreateDelivery :one
INSERT INTO deliveries (event_id, endpoint_url) VALUES ($1, $2) RETURNING *;

-- name: GetPendingDeliveries :many
SELECT * FROM deliveries 
WHERE status = 'pending' AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
ORDER BY next_attempt_at NULLS FIRST
LIMIT 100
FOR UPDATE SKIP LOCKED;

-- name: ChangeStatus :exec
UPDATE deliveries
SET status = $2
WHERE id = $1;

-- name: ProcessFailedAttempt :exec
UPDATE deliveries
SET
    status = CASE
        WHEN attempts + 1 >= sqlc.arg(max_attempts) THEN 'failed'::delivery_status
        ELSE 'pending'::delivery_status
    END,
    attempts = attempts + 1, 
    next_attempt_at = $2,
    last_error = $3
WHERE id = $1;

-- name: ProcessSuccessfulAttempt :exec
UPDATE deliveries
SET
    status = 'delivered',
    last_error = NULL
WHERE id = $1;